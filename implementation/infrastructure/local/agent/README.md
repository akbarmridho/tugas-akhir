# Agent Tester

Prerequisites:

- `sudo apt-get install gettext-base`

## K3d Cluster

Setup cluster:

```bash
k3d cluster create --config ./k3d.yaml
```

Delete cluster:

```bash
k3d cluster delete k6-agent
```

## Build Image

From the `implementation/agent` folder context.

### Build Agent

```bash
docker build -f Dockerfile -t tugas-akhir/agent:latest . &&
docker tag tugas-akhir/agent:latest registry.localhost:5002/tugas-akhir/agent:latest &&
docker push registry.localhost:5002/tugas-akhir/agent:latest
```

## K6 Operator and Monitoring

```bash
helmfile apply
```

### Accessing Grafana

```bash
export POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=grafana" -o jsonpath="{.items[0].metadata.name}") && kubectl --namespace monitoring port-forward $POD_NAME 3000
```

The username is `admin` and the password is `tugas-akhir`.

### Accessing Prometheus

```bash
export POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=prometheus,app.kubernetes.io/instance=prometheus" -o jsonpath="{.items[0].metadata.name}") && kubectl --namespace monitoring port-forward $POD_NAME 9090
```

## Running Test

Required envs:

- `RUN_ID` any randomly generated unique string.
- `VARIANT` with the following values: `smoke`, `smokey`, `sim-1`, `sim-2`, `sim-test`, `stress-1`, `stress-2`, and `stress-test`. This differentiate the k6 agent request pattern and behaviour.

### Start the Test

Before running the test, generate a `RUN_ID`. This is will be used to start and clean up the test.

```bash
openssl rand -hex 6
```

Prepare the env.

```bash
export RUN_ID=<your_ run_id>
export VARIANT=<your_scenario>
export HOST_FORWARD=<host-ip>
```

For local:

```bash
export RUN_ID=local-test
export VARIANT=smokey
export HOST_FORWARD=192.168.65.254
```

To get the host ip, run the following commands:

```bash
# list the monitoring pods
kubectl get pods -n monitoring

# exec to one of the pods
kubectl exec -n monitoring -ti <pod_name> -- bash

# get the resolved ip for host.k3d.internal
nslookup host.k3d.internal

# or host.docker.internal
nslookup host.docker.internal
```

Set the ip returned as the `HOST_FORWARD` value.

Build the code in the `implementation/agent` folder context.

```bash
npm run build
```

Start the test.

```bash
cp ../../../agent/dist/tests/ticket.js ./ticket.js && kubectl create configmap ticket-code --from-file=ticket.js && envsubst < k6.yaml | kubectl apply -f -
```

### Test Cleanup

Ensure that the env used still exist and equal.

```bash
kubectl delete configmap ticket-code && envsubst < k6.yaml | kubectl delete -f -
```
