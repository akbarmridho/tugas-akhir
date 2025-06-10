# Backup Guide

## Call Snapshot

For backend cluster:

```bash
SNAPSHOT_NAME=$(curl -XPOST -k https://prometheus.tugas-akhir.local/api/v1/admin/tsdb/snapshot \
  | jq -r '.data.name')
```

For K6 Agent:

```bash
SNAPSHOT_NAME=$(curl -XPOST -k https://prometheus.k6-agent.local/api/v1/admin/tsdb/snapshot \
  | jq -r '.data.name')
# or
SNAPSHOT_NAME=$(curl -XPOST -k https://prometheus.k6-agent.local:8443/api/v1/admin/tsdb/snapshot \
  | jq -r '.data.name')
```

## Copy the Data

The snapshot result data will be in the following path: `/data/snapshots`.

For each cluster:

```bash
# Example for Kubernetes
POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=prometheus,app.kubernetes.io/instance=prometheus" -o jsonpath="{.items[0].metadata.name}")
mkdir ./backup-data/${SNAPSHOT_NAME}
kubectl cp monitoring/${POD_NAME}:/data/snapshots/${SNAPSHOT_NAME} ./backup-data/${SNAPSHOT_NAME} -c prometheus-server

tar cvzf ${SNAPSHOT_NAME}.tar.gz ./backup-data/${SNAPSHOT_NAME}

curl -F "file=@./${SNAPSHOT_NAME}.tar.gz" https://temp.sh/upload
```
