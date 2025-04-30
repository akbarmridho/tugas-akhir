# Redis Cluster

```bash
chmod +x helmfile
./helmfile apply
```

## Cleanup

```bash
./helmfile delete
kubectl delete pvc data-rabbitmq-0
```
