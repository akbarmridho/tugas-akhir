# Setup Backend Cluster

## Storage Alias

within `infrastructure/simulation` folder, run `kubectl apply -f claim-standard-alias.yaml`.

## Monitoring

Inside the `infrastructure/simulation/monitoring` folder context:

- Inside the `helmfile.yaml` file, check for alloy configuration. Change the cluster name `k3d-tugas-akhir` into the appropriate name. Run `kubectl config get-clusters` to get the cluster name.
- Run `helmfile apply`.

## Nginx

Inside the `infrastructure/simulation/nginx` folder context, run the following commands:

```bash
helmfile apply
kubectl apply -f cert-manager.yaml
kubectl apply -f ingress-payment.yaml
kubectl apply -f ingress-monitoring.yaml
kubectl apply -f ingress-ticket.yaml
```

## Payment Redis Cluster

Inside the `infrastructure/simulation/payment` folder context, run `helmfile apply`.

## TLS

```bash
kubectl create secret tls service-tls \
  --cert=../../../cert/cert.pem \
  --key=../../../cert/key.pem

kubectl create namespace payment

kubectl create secret tls -n payment service-tls \
  --cert=../../../cert/cert.pem \
  --key=../../../cert/key.pem
```

## Additional

Inside `intrastructure/local/postgres` folder context.

```bash
kubectl create secret generic pgbouncer-backend-ca-secret --from-file=pg-ca.pem=certs/ca.pem
kubectl create secret generic pgbouncer-backend-client-cert-secret --from-file=pg-client-cert.crt=certs/client.crt
kubectl create secret generic pgbouncer-backend-client-key-secret --from-file=pg-client-key.key=certs/client.key
kubectl create secret generic pgbouncer-backend-server-cert-secret --from-file=pg-server-cert.crt=certs/server.crt
kubectl create secret generic pgbouncer-backend-server-key-secret --from-file=pg-server-key.key=certs/server.key
```
