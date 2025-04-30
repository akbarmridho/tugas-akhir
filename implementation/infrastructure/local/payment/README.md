# Payment Service

## Build Image

From the `implementation/payment` folder context.

### Build Payment Backend

```bash
docker build -f server.Dockerfile -t tugas-akhir/payment-server:latest .
docker tag tugas-akhir/payment-server:latest registry.localhost:5001/tugas-akhir/payment-server:latest
docker push registry.localhost:5001/tugas-akhir/payment-server:latest
```

### Build Payment Notifier

```bash
docker build -f notifier.Dockerfile -t tugas-akhir/payment-notifier:latest .
docker tag tugas-akhir/payment-notifier:latest registry.localhost:5001/tugas-akhir/payment-notifier:latest
docker push registry.localhost:5001/tugas-akhir/payment-notifier:latest
```

## Setup

Setup the dependencies (Redis Cluster).

```bash
chmod +x helmfile
./helmfile apply
```

Setup the service.

```bash
kubectl apply -f payment.yaml -n payment

## Cleanup

```bash
./helmfile delete
kubectl delete pvc redis-data-redis-redis-cluster-0 -n payment
kubectl delete pvc redis-data-redis-redis-cluster-1 -n payment
kubectl delete pvc redis-data-redis-redis-cluster-2 -n payment
```
