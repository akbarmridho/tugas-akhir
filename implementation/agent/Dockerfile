# Build the k6 binary with the extension
FROM golang:1.24 AS builder

RUN go install go.k6.io/xk6/cmd/xk6@v0.19.1

# For our example, we'll add support for output of test metrics to InfluxDB v2.
# Feel free to add other extensions using the '--with ...'.
RUN xk6 build \
    --with github.com/grafana/xk6-faker@v0.4.3 \
    --output /k6

# Use the operator's base image and override the k6 binary
FROM grafana/k6:1.0.0

COPY --from=builder /k6 /usr/bin/k6