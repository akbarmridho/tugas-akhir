#!/bin/bash

if [[ $UID -ge 10000 ]]; then
    GID=$(id -g)
    sed -e "s/^postgres:x:[^:]*:[^:]*:/postgres:x:$UID:$GID:/" /etc/passwd > /tmp/passwd
    cat /tmp/passwd > /etc/passwd
    rm /tmp/passwd
fi

cat > /home/postgres/patroni.yml <<__EOF__
bootstrap:
  dcs:
    postgresql:
      parameters:
        max_connections: 200
        shared_buffers: 512MB
        ssl: 'on'
        ssl_ca_file: ${PGSSLROOTCERT}
        ssl_cert_file: ${PGSSLCERT}
        ssl_key_file: ${PGSSLKEY}
        shared_preload_libraries: 'pg_stat_statements'
      use_pg_rewind: true
      pg_hba:
      - local all all trust
      - host all all 127.0.0.1/32 trust
      - hostssl replication ${PATRONI_REPLICATION_USERNAME} all md5 clientcert=${PGSSLMODE}
      - hostssl all all all md5
  initdb:
  - auth-host: md5
  - auth-local: trust
  - encoding: UTF8
  - locale: en_US.UTF-8
  - data-checksums
restapi:
  connect_address: '${PATRONI_KUBERNETES_POD_IP}:8008'
postgresql:
  basebackup:
    checkpoint: fast
  connect_address: '${PATRONI_KUBERNETES_POD_IP}:5432'
  authentication:
    superuser:
      sslmode: ${PGSSLMODE}
      sslkey: ${PGSSLKEY}
      sslcert: ${PGSSLCERT}
      sslrootcert: ${PGSSLROOTCERT}
      password: '${PATRONI_SUPERUSER_PASSWORD}'
    replication:
      sslmode: ${PGSSLMODE}
      sslkey: ${PGSSLKEY}
      sslcert: ${PGSSLCERT}
      sslrootcert: ${PGSSLROOTCERT}
      password: '${PATRONI_REPLICATION_PASSWORD}'
__EOF__

unset PATRONI_SUPERUSER_PASSWORD PATRONI_REPLICATION_PASSWORD

exec /usr/bin/python3 /usr/local/bin/patroni /home/postgres/patroni.yml