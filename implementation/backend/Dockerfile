FROM golang:1.24.0-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/usr/local/share \
    GOCACHE=/usr/local/share go mod download

COPY . .

RUN go build -o ./nofc_server ./cmd/nofc/server/main_server.go
RUN go build -o ./fc_worker ./cmd/fc/worker/main_worker.go
RUN go build -o ./fc_server ./cmd/fc/server/main_server.go
RUN go build -o ./seeder ./cmd/seeder/seeder.go

FROM builder AS runtime

WORKDIR /app

COPY ./ca.pem /etc/ssl/pg-ca.pem
COPY ./client.crt /etc/ssl/pg-client-cert.crt
COPY ./client.key /etc/ssl/private/pg-client-key.key

RUN chmod 664 /etc/ssl/pg-ca.pem
RUN chmod 664 /etc/ssl/pg-client-cert.crt
RUN chmod 600 /etc/ssl/private/pg-client-key.key

COPY --from=builder /app/nofc_server /app/nofc_server
COPY --from=builder /app/fc_worker /app/fc_worker
COPY --from=builder /app/fc_server /app/fc_server
COPY --from=builder /app/seeder /app/seeder

EXPOSE 3000
EXPOSE 3001