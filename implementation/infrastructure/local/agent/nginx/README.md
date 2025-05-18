# Nginx Ingress Controller

```bash
chmod +x helmfile
helmfile apply
kubectl apply -f cert-manager.yaml
kubectl apply -f ingress-monitoring.yaml
```

## Cleanup

```bash
kubectl delete -f ingress-monitoring.yaml
kubectl delete -f cert-manager.yaml
helmfile delete
```
