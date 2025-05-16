# Nginx Ingress Controller

```bash
chmod +x helmfile
./helmfile apply
kubectl apply -f cert-manager.yaml
kubectl apply -f ingress-payment.yaml
kubectl apply -f ingress-grafana.yaml
kubectl apply -f ingress-ticket.yaml
```

## Cleanup

```bash
kubectl delete -f ingress-payment.yaml
kubectl delete -f ingress-grafana.yaml
kubectl delete -f ingress-ticket.yaml
kubectl delete -f cert-manager.yaml
./helmfile delete
```
