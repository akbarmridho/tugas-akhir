# Simulation

Read [Test Guide](../test-guide/README.md).

## Resource Planning K6 Agent

Total available resources: 3x 16/32GB with K3s overhead.

| Service Name | Request Allocation | Limit Allocation | Other    |
| ------------ | ------------------ | ---------------- | -------- |
| Nginx        | 0.5/0.5Gi          | 1/1Gi            | -        |
| Cert Manager | 0.1/0.25Gi         | 0.25/0.385Gi     | -        |
| Grafana      | 0.5/0.75Gi         | 1/1.5Gi          | PVC 10Gi |
| Prometheus   | 2/4Gi              | 4/8Gi            | PVC 50Gi |
| K6 Run       | 9 x 4/8Gi          | 9x 4/8Gi         | -        |

## Resource Planning Backend Cluster

### Auxiliary Component

| Service Name    | Request Allocation | Limit Allocation | Other    |
| --------------- | ------------------ | ---------------- | -------- |
| Prometheus      | 0.5/2Gi            | 0.75/4Gi         | PVC 50Gi |
| Alloy           | 0.5/0.25Gi         | 0.75/0.5Gi       | -        |
| Grafana         | 0.25/0.5Gi         | 0.5/0.75Gi       | PVC 10Gi |
| Loki            | 0.5/1.5Gi          | 0.75/2Gi         | -        |
| Nginx           | 2/2Gi              | 3/2.5Gi          | -        |
| Payment Redis   | 3x @ 0.5/0.75Gi    | 3x @ 0.5/0.75Gi  | -        |
| Payment Backend | 1/2Gi              | 1/2Gi            | -        |
| Payment Worker  | 0.5/1Gi            | 0.5/1Gi          | -        |
| Cert Manager    | 0.1/0.25Gi         | 0.25/0.384Gi     | -        |
| PGCat           | 2/1Gi              | 2/1Gi            | -        |
| Ticket Sanity   | 0.25/0.25Gi        | 0.25/0.25Gi      | -        |

### Main Component

**No Flow Control:**

| Service Name               | Postgres No FC | Citus No FC    | YugaByte No FC |
| -------------------------- | -------------- | -------------- | -------------- |
| Postgres Primary & Replica | 2 x 3.75/8Gi   | -              | -              |
| Citusdata Coordinator      | -              | 4.5/6Gi        | -              |
| Citusdata Worker           | -              | 2 x 2/5Gi      | -              |
| YugabyteDB Master          | -              | -              | ?              |
| YugabyteDB TServer         | -              | -              | ?              |
| Redis Cluster              | 3 x 0.75/1.5Gi | 3 x 0.75/1.5Gi | ?              |
| Ticket Backend             | 4 x 2/4Gi      | 4 x 2/4Gi      | ?              |

**Flow Control:**

| Service Name               | Postgres FC | Citus FC | YugaByte FC |
| -------------------------- | ----------- | -------- | ----------- |
| Postgres Primary & Replica | ?           | -        | -           |
| Citusdata Coordinator      | -           | ?        | -           |
| Citusdata Worker           | -           | ?        | -           |
| YugabyteDB Master          | -           | -        | ?           |
| YugabyteDB TServer         | -           | -        | ?           |
| Redis Cluster              | ?           | ?        | ?           |
| RabbitMQ                   | ?           | ?        | ?           |
| Ticket Backend             | ?           | ?        | ?           |
| Ticket Worker              | ?           | ?        | ?           |
