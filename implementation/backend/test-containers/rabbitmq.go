package test_containers

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

const TestRabbitMQUser = "tugas-akhir"
const TestRabbitMQPassword = "tugas-akhir"

type RabbitMQContainer struct {
	testcontainers.Container
}

func (r *RabbitMQContainer) Cleanup(t testing.TB) {
	testcontainers.CleanupContainer(t, r.Container)
}

func NewRabbitMQContainer(ctx context.Context) (*RabbitMQContainer, error) {
	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:4.1.0-management",
		rabbitmq.WithAdminUsername(TestRabbitMQUser),
		rabbitmq.WithAdminPassword(TestRabbitMQPassword),
	)

	if err != nil {
		return nil, err
	}

	return &RabbitMQContainer{
		Container: rabbitmqContainer.Container,
	}, nil
}

func GetRabbitMQContainer(t *testing.T) *RabbitMQContainer {
	ctx := context.Background() // Use background for setup, test context can time out

	container, err := NewRabbitMQContainer(ctx)

	require.NoError(t, err, "Failed to set up Redis cluster")

	// Register cleanup function with the test
	t.Cleanup(func() {
		container.Cleanup(t)
	})

	return container
}
