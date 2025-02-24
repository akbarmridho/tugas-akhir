# Tugas Akhir Backend

## Prerequisites

1. Install `dbmate`

```sh
npm install -g dbmate
```

## Commands

Base paths:

- `/schemas/radar/postgres` For Base Architecture
- `/schemas/pgp/postgres` For Postgres Plus Architecture
- `/schemas/pgp/risingwave` For Postgres Plus Architecture
- `/schemas/eda/risingwave` For EDA Architecture

### Create New Migration

```sh
dbmate -d "{migration_path}/migrations" new <migration-name>
```

### Migrate

```sh
dbmate --env-file ".env" -d "{migration_path}/migrations" -s "{migration_path}/schema.sql" migrate
```

### Rollback

```sh
dbmate --env-file ".env" -d "{migration_path}/migrations" -s "{migration_path}/schema.sql" rollback
```
