# Backup Guide

todo include steps for back up both backend and k6 agent.

## Call Snapshot

```bash
curl -XPOST -k https://prometheus.tugas-akhir.local/api/v1/admin/tsdb/snapshot
```

For K6 Agent

```bash
curl -XPOST -k https://prometheus.k6-agent.local/api/v1/admin/tsdb/snapshot
# or
curl -XPOST -k https://prometheus.k6-agent.local:8443/api/v1/admin/tsdb/snapshot
```

## Copy the Data

The snapshot result data will be in the following path: `/data/snapshots`.

## Copy the Snapshot

```bash
# Example for Kubernetes
POD_NAME=$(kubectl get pods -l <your-prometheus-label-selector> -o jsonpath='{.items[0].metadata.name}')
SNAPSHOT_NAME="<output_from_curl_command>" # e.g., 20250517T103000Z-abcdef1234567890
kubectl cp <namespace>/${POD_NAME}:/data/snapshots/${SNAPSHOT_NAME} ./<local-backup-path>/${SNAPSHOT_NAME} -c <prometheus-container-name>
```
