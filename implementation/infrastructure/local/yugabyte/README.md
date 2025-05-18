# YugaByteDB

```bash
chmod +x helmfile
helmfile apply
```

## Cleanup

```bash
helmfile delete
kubectl delete pvc --namespace default -l app=yb-master
kubectl delete pvc --namespace default -l app=yb-tserver
```
