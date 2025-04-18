package test_containers

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"tugas-akhir/backend/infrastructure/config"
	redisInfra "tugas-akhir/backend/infrastructure/redis"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	redisImage     = "redis:7.4"
	redisPort      = "6379/tcp"
	clusterBusPort = "16379/tcp"
	nodeCount      = 3
)

type RedisCluster struct {
	Containers []testcontainers.Container
	Network    *testcontainers.DockerNetwork
	MappedAddr []string // Stores ["host:port", "host:port", ...] for client connection
	AliasAddr  []string
}

func (rc *RedisCluster) Cleanup(t testing.TB) {
	ctx := context.Background() // Use background context for cleanup
	log.Println("Cleaning up Redis Cluster...")
	for i := len(rc.Containers) - 1; i >= 0; i-- {
		if rc.Containers[i] != nil {
			log.Printf("Terminating Redis container %d...", i+1)
			if err := rc.Containers[i].Terminate(ctx); err != nil {
				t.Errorf("Failed to terminate Redis container %d: %v", i+1, err)
			}
		}
	}
	if rc.Network != nil {
		log.Println("Removing network...")
		if err := rc.Network.Remove(ctx); err != nil {
			t.Errorf("Failed to remove network: %v", err)
		}
	}
	log.Println("Redis Cluster cleanup finished.")
}

// NewRedisCluster sets up a 3-node Redis cluster
func NewRedisCluster(ctx context.Context) (*RedisCluster, error) {
	cluster := &RedisCluster{
		Containers: make([]testcontainers.Container, nodeCount),
		MappedAddr: make([]string, nodeCount),
	}

	// 1. Create a new Docker network for the cluster nodes
	//log.Println("Creating Docker network for Redis Cluster...")
	net, err := network.New(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	cluster.Network = net
	networkName := net.Name
	//log.Printf("Network '%s' created.", networkName)

	// Prepare node addresses for the cluster create command (using internal network aliases)
	nodeAddrsInternal := make([]string, nodeCount)
	nodeAliases := make([]string, nodeCount)

	// 2. Define and start each Redis node container
	for i := 0; i < nodeCount; i++ {
		nodeAliases[i] = fmt.Sprintf("redis-node-%d", i+1)
		// Internal address will be like "redis-node-1:6379"
		nodeAddrsInternal[i] = fmt.Sprintf("%s:%s", nodeAliases[i], nat.Port(redisPort).Port())

		networkAlias := make(map[string][]string)

		networkAlias[networkName] = []string{nodeAliases[i]}

		req := testcontainers.ContainerRequest{
			Image:          redisImage,
			ExposedPorts:   []string{redisPort, clusterBusPort},
			NetworkAliases: networkAlias,
			Networks:       []string{networkName},
			Cmd: []string{
				"redis-server",
				"--port", nat.Port(redisPort).Port(),
				"--cluster-enabled", "yes",
				"--cluster-config-file", fmt.Sprintf("/data/nodes-%d.conf", i+1), // Unique config file per node
				"--cluster-node-timeout", "5000",
				"--appendonly", "yes",
				// Important for Docker: Announce the node's alias or IP within the Docker network
				// Using alias is generally more reliable than trying to guess internal IP
				"--cluster-announce-ip", nodeAliases[i], // Use alias for announcement
				"--cluster-announce-port", nat.Port(redisPort).Port(),
				"--cluster-announce-bus-port", nat.Port(clusterBusPort).Port(),
			},
			WaitingFor: wait.ForLog("Ready to accept connections").WithStartupTimeout(20 * time.Second),
		}

		//log.Printf("Starting Redis container %d (%s)...", i+1, nodeAliases[i])
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			for j := 0; j < i; j++ {
				_ = cluster.Containers[j].Terminate(ctx)
			}
			_ = net.Remove(ctx)
			return nil, fmt.Errorf("failed to start container %d (%s): %w", i+1, nodeAliases[i], err)
		}
		cluster.Containers[i] = container
		//log.Printf("Redis container %d (%s) started.", i+1, nodeAliases[i])

		// Get mapped host and port for client connection
		host, err := container.Host(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get host for container %d: %w", i+1, err)
		}
		mappedPort, err := container.MappedPort(ctx, redisPort)
		if err != nil {
			return nil, fmt.Errorf("failed to get mapped port for container %d: %w", i+1, err)
		}
		cluster.MappedAddr[i] = fmt.Sprintf("%s:%s", host, mappedPort.Port())
		//log.Printf("Container %d (%s) accessible at: %s", i+1, nodeAliases[i], cluster.MappedAddr[i])
	}

	// 3. Create the cluster using redis-cli on the first node
	//log.Println("All Redis nodes started. Attempting cluster creation...")

	// Build the cluster create command arguments dynamically
	clusterCreateCmdArgs := []string{"redis-cli", "--cluster", "create"}
	clusterCreateCmdArgs = append(clusterCreateCmdArgs, nodeAddrsInternal...)
	clusterCreateCmdArgs = append(clusterCreateCmdArgs, "--cluster-replicas", "0", "--cluster-yes") // 0 replicas = all masters

	//log.Printf("Executing cluster create command on node 1: %v", clusterCreateCmdArgs)

	// Execute the command within the first container
	exitCode, cmdOutput, err := cluster.Containers[0].Exec(ctx, clusterCreateCmdArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cluster create command: %w", err)
	}

	// Read output
	outputBytes, readErr := io.ReadAll(cmdOutput)
	outputString := string(outputBytes)
	//log.Printf("Cluster create command output (Exit Code: %d):\n%s", exitCode, outputString)

	cluster.AliasAddr = nodeAliases

	for i, addr := range cluster.AliasAddr {
		cluster.AliasAddr[i] = fmt.Sprintf("%s:%s", addr, nat.Port(redisPort).Port())
	}

	// Check if the command was successful
	if exitCode != 0 || !strings.Contains(outputString, "[OK] All 16384 slots covered.") {
		if readErr != nil { // Include potential read error
			return nil, fmt.Errorf("cluster create command failed with exit code %d and read error %v. Output: %s", exitCode, readErr, outputString)
		}
		return nil, fmt.Errorf("cluster create command failed with exit code %d. Output: %s", exitCode, outputString)
	}

	log.Println("Redis Cluster created successfully!")
	return cluster, nil
}

// GetRedisCluster sets up the 3-node cluster and returns a connected client
func GetRedisCluster(t *testing.T) *redisInfra.Redis {
	ctx := context.Background() // Use background for setup, test context can time out

	clusterSetupTimeout := 2 * time.Minute // Increase timeout for multi-node setup
	setupCtx, cancel := context.WithTimeout(ctx, clusterSetupTimeout)
	defer cancel()

	//log.Println("Setting up 3-node Redis Cluster...")
	cluster, err := NewRedisCluster(setupCtx)
	require.NoError(t, err, "Failed to set up Redis cluster")

	// Register cleanup function with the test
	t.Cleanup(func() {
		cluster.Cleanup(t)
	})

	//log.Println("Redis cluster nodes mapped addresses:", cluster.MappedAddr)

	cfg := config.Config{
		RedisHosts:    strings.Join(cluster.MappedAddr, ","),
		RedisHostsMap: strings.Join(cluster.AliasAddr, ","),
	}

	//log.Println("Connecting Redis client to cluster...")
	redisConn, err := redisInfra.NewRedis(&cfg)
	require.NoError(t, err, "Failed to create Redis client")

	// Test client connection
	log.Println("Checking Redis client health...")

	// add some time to wair for redis cluster to be healthy
	time.Sleep(10 * time.Second)

	// Use test's context for health check if appropriate, or background if needed
	err = redisConn.IsHealthy(ctx)
	require.NoError(t, err, "Redis client health check failed")
	log.Println("Redis client connected and healthy.")

	return redisConn
}
