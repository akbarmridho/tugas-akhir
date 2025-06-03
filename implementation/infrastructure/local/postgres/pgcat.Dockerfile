FROM ghcr.io/postgresml/pgcat:4a7a6a8e7a78354b889002a4db118a8e2f2d6d79

COPY certs/ca.pem /etc/ssl/pg-ca.pem
COPY certs/server.crt /etc/ssl/pg-server-cert.crt
COPY certs/server.key /etc/ssl/private/pg-server-key.key

RUN chmod 664 /etc/ssl/pg-ca.pem
RUN chmod 664 /etc/ssl/pg-server-cert.crt
RUN chmod 600 /etc/ssl/private/pg-server-key.key
