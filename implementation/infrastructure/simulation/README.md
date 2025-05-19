# Simulation

Read [Test Guide](../test-guide/README.md).

## Resource Planning K6 Agent

Total available resources: 3x 16/32GB with K3s overhead.

| Service Name | Request Allocation | Limit Allocation | Other    |
| ------------ | ------------------ | ---------------- | -------- |
| Nginx        | 0.5/0.5Gi          | 1/1Gi            | -        |
| Grafana      | 0.5/0.75Gi         | 1/1.5Gi          | PVC 10Gi |
| Prometheus   | 2/4Gi              | 4/8Gi            | PVC 50Gi |
| K6           | 3x @ 13/26Gi       | 3x @ 13/26Gi     | -        |

Total CPU request: 42 out of 48 core.
Total RAM request: 83 out of 96 GB.

## Resource Planning Backend Cluster

### Auxiliary Component

| Service Name    | Request Allocation | Limit Allocation | Other    |
| --------------- | ------------------ | ---------------- | -------- |
| Prometheus      | 2/4G               | 2.5/4.5G         | PVC 50Gi |
| Alloy           | 0.5/1Gi            | 0.75/1.25Gi      | -        |
| Grafana         | 0.5/0.75Gi         | 1/1.5Gi          | PVC 10Gi |
| Loki            | 1.5/3Gi            | 2/3.5Gi          | -        |
| Nginx           | 2.5/2Gi            | 3.5/2.5Gi        | -        |
| Payment Redis   | 3x @ 0.5/1Gi       | 3x @ 0.75/1.25Gi | -        |
| Payment Backend | 1/2Gi              | 1.25/2.5Gi       | -        |
| Payment Worker  | 1/2Gi              | 1.25/2.5Gi       | -        |

Total CPU request: 10.5 CPU.
Total RAM request: 17.75 GB RAM.

### Main Component

| Service Name             | Postgres No FC | Citus No FC | YugaByte No FC | Citus FC | YugaByte FC |
| ------------------------ | -------------- | ----------- | -------------- | -------- | ----------- |
| Postgres Primary         | 3.5/8Gi        | -           | -              | -        | -           |
| Postgres Replica         | 3.5/8Gi        | -           | -              | -        | -           |
| Citusdata Node 1 (Coord) | -              | 4/8Gi       | -              | 3.5/7Gi  | -           |
| Citusdata Node 2         | -              | 3/8Gi       | -              | 2.5/7Gi  | -           |
| YugabyteDB Node 1        | -              | -           | 3.5/8Gi        | -        | 3/7Gi       |
| YugabyteDB Node 2        | -              | -           | 3.5/8Gi        | -        | 3/7Gi       |
| Redis Cluster (total)    | 3/4.5Gi        | 3/4.5Gi     | 3/4.5Gi        | 3/6Gi    | 3/6Gi       |
| RabbitMQ                 | -              | -           | -              | 1.5/3Gi  | 1.5/3Gi     |
| Ticket Backend (total)   | 11/24Gi        | 11/24Gi     | 11/24Gi        | 8.5/17Gi | 8.5/17Gi    |
| Ticket Worker (total)    | -              | -           | -              | 2/4Gi    | 2/4Gi       |

For YugaByteDB, 0.5 CPU and 1Gi RAM will be allocated for master and the rest will be allocated to tserver.

Postgres No FC total CPU: 21 CPU.
Postgres No FC total RAM: 44.5 GB.

CitusData No FC total CPU: 21 CPU.
CitusData No FC total RAM: 44.5 GB.

YugaByte No FC total CPU: 21 CPU.
YugaByte No FC total RAM: 44.5 GB.

CitusData FC total CPU: 21 CPU.
CitusData FC total RAM: 44 GB.

YugaByte FC total CPU: 21 CPU.
YugaByte FC total RAM: 44 GB.
