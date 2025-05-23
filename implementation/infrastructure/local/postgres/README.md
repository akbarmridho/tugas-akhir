# Postgres with Patroni

Original source: [Patroni](https://github.com/patroni/patroni/blob/master/kubernetes/README.md)

## Postgres Steps

### Build Postgres Docker Image

```bash
docker build -f postgres.Dockerfile -t tugas-akhir/postgres:latest . &&
docker tag tugas-akhir/postgres:latest registry.localhost:5001/tugas-akhir/postgres:latest &&
docker push registry.localhost:5001/tugas-akhir/postgres:latest
```

### Client Secret

```bash
kubectl create secret generic pgbouncer-backend-ca-secret --from-file=pg-ca.pem=certs/ca.pem
kubectl create secret generic pgbouncer-backend-client-cert-secret --from-file=pg-client-cert.crt=certs/client.crt
kubectl create secret generic pgbouncer-backend-client-key-secret --from-file=pg-client-key.key=certs/client.key
kubectl create secret generic pgbouncer-backend-server-cert-secret --from-file=pg-server-cert.crt=certs/server.crt
kubectl create secret generic pgbouncer-backend-server-key-secret --from-file=pg-server-key.key=certs/server.key
```

### Apply Postgres Kubernetes

```bash
kubectl apply -f postgres.yaml
helmfile apply -f helmfile-postgres.yaml
```

### Delete Postgres Kubernetes

```bash
kubectl delete -f postgres.yaml
helmfile delete -f helmfile-postgres.yaml
kubectl delete svc pgcluster-config
```

### Temporary Access

```bash
kubectl port-forward pod/pgcluster-0 5432:5432
```

### Check

**Check placement:**

```bash
kubectl get pods -o wide -l application=postgres
```

**Check patronictl:**

```bash
kubectl exec -ti pgcluster-0 -- bash
patronictl list
```

## Citus Steps

### Build Citus Docker Image

```bash
docker build -f citus.Dockerfile -t tugas-akhir/citus:latest . &&
docker tag tugas-akhir/citus:latest registry.localhost:5001/tugas-akhir/citus:latest &&
docker push registry.localhost:5001/tugas-akhir/citus:latest
```

### Apply Citus Kubernetes

```bash
kubectl apply -f citus.yaml
helmfile apply -f helmfile-citus.yaml
```

### Delete Citus Kubernetes

```bash
kubectl delete -f citus.yaml
helmfile delete -f helmfile-citus.yaml
kubectl delete svc cituscluster-0-config cituscluster-1-config cituscluster-2-config
kubectl delete endpoints cituscluster-0-sync cituscluster-1-sync cituscluster-2-sync
```

### Check

```bash
kubectl get pods -o wide -l cluster-name=cituscluster -L role
kubectl exec -ti cituscluster-0-0 -- bash
psql citus
table pg_dist_node;
```
