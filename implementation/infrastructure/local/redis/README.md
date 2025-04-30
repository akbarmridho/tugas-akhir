# Redis Cluster

```bash
chmod +x helmfile
./helmfile apply
```

## Cleanup

```bash
./helmfile delete
kubectl delete pvc redis-data-redis-redis-cluster-0
kubectl delete pvc redis-data-redis-redis-cluster-1
kubectl delete pvc redis-data-redis-redis-cluster-2
```
