# Debug

## Profiling

Ensure env for profiling is enabled.

```bash
curl -o trace.out -k https://ticket.tugas-akhir.local/debug/pprof/trace?seconds=30
curl -o citus-baru.out -k https://ticket.tugas-akhir.local/debug/pprof/trace?seconds=30

go tool trace trace.out

curl -o block.out -k https://ticket.tugas-akhir.local/debug/pprof/block?seconds=30

go tool pprof -web nofc_server block.out
```

## Postgres Debug

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

--- postgres
SELECT query, calls, total_exec_time, mean_exec_time
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

--- yugabyte
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

select count(*), state
from pg_stat_activity
group by state;

LOAD 'auto_explain';
SET auto_explain.log_min_duration = '100ms';
SET auto_explain.log_analyze = true;
SET auto_explain.log_verbose = true;
```
