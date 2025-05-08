# Nginx Ingress Controller

```bash
chmod +x helmfile
./helmfile apply
kubectl apply -f cert-manager.yaml
kubectl apply -f ingress-payment.yaml
kubectl apply -f ingress-ticket.yaml
```

## Cleanup

```bash
./helmfile delete
kubectl delete -f cert-manager.yaml
kubectl delete -f ingress-payment.yaml
kubectl delete -f ingress-ticket.yaml
```
