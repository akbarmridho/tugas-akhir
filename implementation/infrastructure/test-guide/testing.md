# Testing

## Preparing Backend Environment

For **No Flow Control** case, setup the following: `Payment Service` and one of the following databases: `PostgreSQL`, `CitusData`, and `YugaByteDB`.

For **Flow Control** case, setup the following: `Payment Service`, `RabbitMQ`, and one of the following databases: `CitusData` and `YugaByteDB`.

### Payment Service

#### Setup

Inside the `infrastructure/simulation/payment` folder context, run the following commands:

```bash
helmfile apply

# run the reset
kubectl apply -f payment-reset.yaml -n payment

# wait for finish
kubectl delete -f payment-reset.yaml -n payment

# setup the instance
kubectl apply -f payment.yaml -n payment
```

#### Teardown

Insire the `infrastructure/simulation/payment folder context, run the following command:

```bash
kubectl delete -f payment.yaml -n payment
```

### Ticket Redis Cluster

#### Setup

Inside the `infrastructure/simulation/redis` folder context, run the following commands:

```bash
helmfile apply -f helmfile-nofc.yaml
# or
helmfile apply -f helmfile-fc.yaml
```

#### Teardown

Insire the `infrastructure/simulation/redis` folder context, run the following command:

```bash
helmfile delete -f helmfile-nofc.yaml
# or
helmfile delete -f helmfile-fc.yaml

kubectl delete pvc redis-data-redis-redis-cluster-0
kubectl delete pvc redis-data-redis-redis-cluster-1
kubectl delete pvc redis-data-redis-redis-cluster-2
```

### PostgreSQL

#### Setup

Inside the `infrastructure/simulation/postgres` folder context, run the following commands:

```bash
kubectl apply -f postgres-nofc.yaml
# or
kubectl apply -f postgres-fc.yaml
```

#### Teardown

Insire the `infrastructure/simulation/postgres` folder context, run the following command:

```bash
kubectl delete -f postgres-nofc.yaml
# or
kubectl delete -f postgres-fc.yaml
```

### CitusData

#### Setup

Inside the `infrastructure/simulation/postgres` folder context, run the following commands:

```bash
kubectl apply -f citus-nofc.yaml
# or
kubectl apply -f citus-fc.yaml
```

#### Teardown

Inside the `infrastructure/simulation/postgres` folder context, run the following command:

```bash
kubectl delete -f citus-nofc.yaml
# or
kubectl delete -f citus-fc.yaml
```

### YugaByteDB

#### Setup

Inside the `infrastructure/simulation/yugabyte` folder context, run the following commands:

```bash
helmfile apply -f helmfile-nofc.yaml
# or
helmfile apply -f helmfile-fc.yaml
```

#### Teardown

Inside the `infrastructure/simulation/yugabyte` folder context, run the following command:

```bash
helmfile delete -f helmfile-nofc.yaml
# or
helmfile delete -f helmfile-fc.yaml

kubectl delete pvc --namespace default -l app=yb-master
kubectl delete pvc --namespace default -l app=yb-tserver
```

### RabbitMQ

#### Setup

Inside the `infrastructure/simulation/rabbitmq` folder context, run the following commands:

```bash
helmfile apply
```

#### Teardown

Inside the `infrastructure/simulation/rabbitmq` folder context, run the following command:

```bash
helmfile delete
kubectl delete pvc data-rabbitmq-0
```

## Running the Test

### Setup - Environment Variable

- For PostgreSQL cluster.

```bash
# direct connection
export DB_VARIANT=postgres
export DATABASE_URL="postgresql://postgres:zalando@pgcluster.default.svc.cluster.local,pgcluster-repl.default.svc.cluster.local:5432/postgres?target_session_attrs=read-write&sslmode=verify-ca&sslrootcert=/etc/ssl/pg-ca.pem&sslcert=/etc/ssl/pg-client-cert.crt&sslkey=/etc/ssl/private/pg-client-key.key&pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=postgres
export DATABASE_URL="postgresql://postgres:zalando@pgbouncer.default.svc.cluster.local,pgpool-pgbouncer.default.svc.cluster.local:5432/postgres?target_session_attrs=read-write&pool_max_conns=1000&pool_min_conns=1"
```

- For Citusdata cluster.

```bash
# direct connection
export DB_VARIANT=citusdata
export DATABASE_URL="postgresql://postgres:zalando@cituscluster-0.default.svc.cluster.local:5432/citus?sslmode=verify-ca&sslrootcert=/etc/ssl/pg-ca.pem&sslcert=/etc/ssl/pg-client-cert.crt&sslkey=/etc/ssl/private/pg-client-key.key&pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=citusdata
export DATABASE_URL="postgresql://postgres:zalando@pgbouncer.default.svc.cluster.local:5432/citus?pool_max_conns=2500&pool_min_conns=1"
```

- For YugabyteDB cluster.

```bash
# direct connection
export DB_VARIANT=yugabytedb
export DATABASE_URL="postgresql://yugabyte@yb-tserver-0.yb-tservers.default.svc.cluster.local:5433,yb-tserver-1.yb-tservers.default.svc.cluster.local:5433/yugabyte?pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=yugabytedb
export DATABASE_URL="postgresql://yugabyte:yugabyte@pgbouncer.default.svc.cluster.local:5432/yugabyte?pool_max_conns=2500&pool_min_conns=1&sslmode=disable"
```

### Seed the Data

**Reminder: ensure that you have reset the Payment Redis Cluster, RabbitMQ, and K6 test logs.**

Inside the `infrastructure/simulation/ticket` folder context:

```bash
export TEST_SCENARIO=<your_scenario>
```

Supported scenarios: `sf-4`, `sf-2`, `sf-1`, `s2-4`, `s2-2`, `s2-1`, `s5-4`, `s5-2`, `s5-1`, `s10-4`, `s10-2`, `s10-1`.

Then run the seed.

**Without Flow Control:**

```bash
export SEED_DROPPER=no
envsubst < ticket-seeder.yaml | kubectl apply -f -
```

**With Flow Control:**

```bash
export SEED_DROPPER=yes
envsubst < ticket-seeder.yaml | kubectl apply -f -
```

Check the logs.

```bash
kubectl logs job/ticket-seeder -f
```

Cleanup the logs.

```bash
kubectl delete job ticket-seeder
```

### Setup Ticket Service

Inside the `infrastructure/simulation/ticket` folder context:

```bash
envsubst < ticket-nofc.yaml | kubectl apply -f -
# or for async case
envsubst < ticket-fc.yaml | kubectl apply -f -
```

For teardown

```bash
envsubst < ticket-nofc.yaml | kubectl delete -f -
# or for async case
envsubst < ticket-fc.yaml | kubectl delete -f -
```

### Setup K6 Test

**Note: check the Grafana dashboard for everything and ensure that everything is up.**

Prepare the env.

```bash
export RUN_ID=<your_ run_id>
export VARIANT=<your_scenario>
export HOST_FORWARD=<host-ip>
```

**Note: Fill in the `HOST_FORWARD` value with the Backend Cluster load balancer IP.**

Available `VARIANT` values: `smoke`, `smokey`, `sim-1`, `sim-2`, `sim-test`, `stress-1`, `stress-2`, and `stress-test`. This differentiate the k6 agent request pattern and behaviour.

Ensure to build the test script from the `implementation/agent` folder context, then running `npm run build`.

Inside the `infrastructure/simulation/agent` folder context:

```bash
# run the test
cp ../../../agent/dist/tests/ticket.js ./ticket.js && kubectl create configmap ticket-code --from-file=ticket.js && envsubst < k6.yaml | kubectl apply -f -
```

Writing the test time (with the format) -> how to write run id?

Write down the test data with the following format:

| Run ID  | Test Variant | Scenario | Flow Control | Start Time             | End Time               |
| ------- | ------------ | -------- | ------------ | ---------------------- | ---------------------- |
| `anyid` | `sim-1`      | `s2-2`   | `yes/no`     | `2025-05-18T12:00:00Z` | `2025-05-18T13:00:00Z` |

**Note: the start time and end time should be in UTC time.**

## Post Test Cleanup

- Wait for the test to finish.
- Check for the Grafana dashboard in Backend cluster and K6 agent cluster.
- Teardown the k6 run by running the following command inside the `infrastructure/simulation/agent` folder context: `kubectl delete configmap ticket-code && envsubst < k6.yaml | kubectl delete -f -`
- Teardown the ticket service by running the following command inside the `infrastructure/simulation/ticket` folder context: `envsubst < ticket-nofc.yaml | kubectl delete -f -` or `envsubst < ticket-fc.yaml | kubectl delete -f -`
