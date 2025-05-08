# Payment Service

## Build Image

From the `implementation/payment` folder context.

```bash
docker build -f Dockerfile -t tugas-akhir/payment:latest .
docker tag tugas-akhir/payment:latest registry.localhost:5001/tugas-akhir/payment:latest
docker push registry.localhost:5001/tugas-akhir/payment:latest
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
```

## Cleanup

```bash
./helmfile delete
kubectl delete -f payment.yaml -n payment
kubectl delete pvc redis-data-redis-redis-cluster-0 -n payment
kubectl delete pvc redis-data-redis-redis-cluster-1 -n payment
kubectl delete pvc redis-data-redis-redis-cluster-2 -n payment
```
