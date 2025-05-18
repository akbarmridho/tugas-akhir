# Setup K6 Agent Cluster

## Storage Alias

within `infrastructure/simulation` folder, run `kubectl apply -f claim-standard-alias.yaml`.

## Monitoring & K6 Operator

Inside the `infrastructure/simulation/agent` folder context, run `helmfile apply`.

## Nginx

Inside the `infrastructure/simulation/agent/nginx` folder context, run the following commands:

```bash
helmfile apply
kubectl apply -f cert-manager.yaml
kubectl apply -f ingress-monitoring.yaml
```
