FROM golang:1.24.0-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/usr/local/share \
    GOCACHE=/usr/local/share go mod download

COPY . .

RUN go build -o ./seeder ./cmd/seeder/seeder.go

FROM builder AS runtime

WORKDIR /app

COPY --from=builder /app/seeder /app/seeder

CMD ["./seeder"]