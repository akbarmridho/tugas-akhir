# Ticket Service

Prerequisites:

- `sudo apt-get install gettext-base`

## Build Image

From the `implementation/backend` folder context.

### Build Ticket Backend

```bash
docker build -f Dockerfile -t tugas-akhir/ticket-server:latest . &&
docker tag tugas-akhir/ticket-server:latest registry.localhost:5001/tugas-akhir/ticket-server:latest &&
docker push registry.localhost:5001/tugas-akhir/ticket-server:latest
```

## Setup

### Prepare Database URL

- For PostgreSQL cluster.

```bash
# direct connection
export DB_VARIANT=postgres
export DATABASE_URL="postgresql://postgres:zalando@pgcluster.default.svc.cluster.local,pgcluster-repl.default.svc.cluster.local:5432/postgres?target_session_attrs=read-write&sslmode=verify-ca&sslrootcert=/etc/ssl/pg-ca.pem&sslcert=/etc/ssl/pg-client-cert.crt&sslkey=/etc/ssl/private/pg-client-key.key&pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=postgres
export DATABASE_URL="postgresql://postgres:zalando@pgbouncer.default.svc.cluster.local,pgpool-pgbouncer.default.svc.cluster.local:5432/postgres?target_session_attrs=read-write&pool_max_conns=2500&pool_min_conns=1"
```

- For Citusdata cluster.

```bash
# direct connection
export DB_VARIANT=citusdata
export DATABASE_URL="postgresql://postgres:zalando@cituscluster-0.default.svc.cluster.local:5432/citus?sslmode=verify-ca&sslrootcert=/etc/ssl/pg-ca.pem&sslcert=/etc/ssl/pg-client-cert.crt&sslkey=/etc/ssl/private/pg-client-key.key&pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=citusdata
export DATABASE_URL="postgresql://postgres:zalando@pgbouncer.default.svc.cluster.local:5432/citus?pool_max_conns=200&pool_min_conns=1"
```

- For YugabyteDB cluster.

```bash
# direct connection
export DB_VARIANT=yugabytedb
export DATABASE_URL="postgresql://yugabyte@yb-tserver-0.yb-tservers.default.svc.cluster.local:5433,yb-tserver-1.yb-tservers.default.svc.cluster.local:5433/yugabyte?pool_max_conns=40&pool_min_conns=1"

# pooled connection
export DB_VARIANT=yugabytedb
export DATABASE_URL="postgresql://yugabyte:yugabyte@pgbouncer.default.svc.cluster.local:5432/yugabyte?pool_max_conns=250&pool_min_conns=1&sslmode=disable"
```

### Prepare the Scenario

```bash
export TEST_SCENARIO=<your_scenario>
```

Supported scenarios: `sf-4`, `sf-2`, `sf-1`, `s2-4`, `s2-2`, `s2-1`, `s5-4`, `s5-2`, `s5-1`, `s10-4`, `s10-2`, `s10-1`.

List of scenario
xx-y
xx variant: sf (scale full), s2 (scale by 2), s3 (scale by 3), ...
y variant: 1, 2, 3 (day count)
Festival/ free seating area can hold 20.000 person.
Lower seat can hold 25.000 person.
Upper seat can hold 35.000 person.

In GBK, lower seat divided into:

- Platinum East 1, Platinum East 2, Platinum West 1, Platinum West 2 @2000 seat -> 1 area
- Gold East 1, Gold East 2, Gold West 1, Gold West 2 @1750 seat -> 1 area
- Silver North, Silver South @5000 seat -> 5 area
  
Upper seat can divided into:

- Bronze North, Bronze South @7000 seat -> 7 area
- Bronze West, Bronze East @10500 seat -> 10 area

Festival can be divided into:

- VIP Total 4000 seat.
- Zone A Total 8000 seat.
- Zone B Total 8000 seat.

### Seeding the Database

**Reset RabbitMQ:**

If using flow control.

```bash
kubectl exec -ti rabbitmq-0 -- bash

# inside the rabbitmq terminal
rabbitmqctl stop_app && rabbitmqctl reset && rabbitmqctl start_app && exit
```

**Reset Payment Service:**

Inside the `implementation/infrastructure/payment`/

```bash
# teardown the payment service
kubectl delete -f payment.yaml -n payment

# run the reset
kubectl apply -f payment-reset.yaml -n payment

# wait for finish
kubectl delete -f payment-reset.yaml -n payment

# up again
kubectl apply -f payment.yaml -n payment
```

Setup the seeder job.

**No Flow Control:**

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

### Setting up the Service

Setup the service.

```bash
envsubst < ticket-nofc.yaml | kubectl apply -f -
# or for async case
envsubst < ticket-fc.yaml | kubectl apply -f -
```

### Deleting the Service

Delete the service.

```bash
envsubst < ticket-nofc.yaml | kubectl delete -f -
# or for async case
envsubst < ticket-fc.yaml | kubectl delete -f -
```
