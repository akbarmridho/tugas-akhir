FROM golang:1.24.0-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/usr/local/share \
    GOCACHE=/usr/local/share go mod download

COPY . .

RUN go build -o ./server ./cmd/fc/server/main_server.go

FROM builder AS runtime

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 3000

CMD ["./server"]