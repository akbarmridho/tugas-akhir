# Simulation

todo akbar:
adjust resource limit
adjust semua path build image.
adjust semua pvc size

Read [Test Guide](../test-guide/README.md).

## Resource Planning K6 Agent

| Service Name | Request Allocation | Limit Allocation | Other      |
| ------------ | ------------------ | ---------------- | ---------- |
| Nginx        | `0.5/0.5Gi`        | `1/1Gi`          | -          |
| Grafana      | `0.5/0.75Gi`       | `1/1.5Gi`        | `PVC 10Gi` |
| Prometheus   | `2/4Gi`            | `4/8Gi`          | `PVC 50Gi` |
| K6           | `3x 12/24Gi`       | `3x 12/24Gi`     | -          |
