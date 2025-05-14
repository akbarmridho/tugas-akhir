# Postgres with Patroni

Original source: [Patroni](https://github.com/patroni/patroni/blob/master/kubernetes/README.md)

## Postgres Steps

### Build Postgres Docker Image

```bash
docker build -f postgres.Dockerfile -t tugas-akhir/postgres:latest . &&
docker tag tugas-akhir/postgres:latest registry.localhost:5001/tugas-akhir/postgres:latest &&
docker push registry.localhost:5001/tugas-akhir/postgres:latest
```

### Apply Postgres Kubernetes

```bash
kubectl apply -f postgres.yaml
```

### Delete Postgres Kubernetes

```bash
kubectl delete -f postgres.yaml
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
```

### Delete Citus Kubernetes

```bash
kubectl delete -f citus.yaml
```

### Check

```bash
kubectl get pods -o wide -l cluster-name=cituscluster -L role
kubectl exec -ti cituscluster-0-0 -- bash
psql citus
table pg_dist_node;
```
