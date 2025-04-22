# Payment Service

## Dev

```bash
pnpm install
npm run dev
```

## Curl Commands

### Home

```bash
curl --http2 -k https://localhost:3000
```

### Metrics

```bash
curl --http2 -k https://localhost:3000/metrics
```

```bash
curl --http2 -k https://localhost:3000/metrics-queue
```

### Health

```bash
curl --http2 -k https://localhost:3000/health
```

### Create Invoice

```bash
curl --http2 -k  'https://localhost:3000/invoices' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data '{
  "amount": 1700000,
  "externalId": "external-id",
  "description": "Yoasobi Concert Tour"
}'
```

### Get Invoice

```bash
curl --http2 -k 'https://localhost:3000/invoices/{id}' \
--header 'Accept: application/json'
```

### Pay Invoice

```bash
curl --http2 -k 'https://localhost:3000/invoices/{id}/payment' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data '{
  "mode": "success"
}'
```

```bash
curl --http2 -k 'https://localhost:3000/invoices/{id}/payment' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data '{
  "mode": "failed"
}'
```
