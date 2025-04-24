# Postgres with Patroni

Original source: [Patroni](https://github.com/patroni/patroni/blob/master/kubernetes/README.md)

## Postgres Steps

### Build Docker Image

```bash
docker build -f postgres.Dockerfile -t tugas-akhir/postgres:latest .
docker tag tugas-akhir/postgres:latest registry.localhost:5001/tugas-akhir/postgres:latest
docker push registry.localhost:5001/tugas-akhir/postgres:latest
```

### Apply Kubernetes

```bash
kubectl apply -f postgres.yaml
```

### Temporary Access

```bash
kubectl port-forward pod/pgcluster-0 5432:5432
```