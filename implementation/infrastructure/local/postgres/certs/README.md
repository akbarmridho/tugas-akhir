# CERTS

## Generate Root CA

```bash
openssl req -x509 -newkey rsa:2048 -days 3650 -nodes \
  -keyout ca.key -out ca.pem -config openssl.cnf
```

## Server Certificate

```bash
openssl req -new -nodes \
  -newkey rsa:2048 \
  -keyout server.key \
  -out server.csr \
  -config server.cnf
```

## Client Certificate

```bash
openssl req -new -nodes \
  -newkey rsa:2048 \
  -keyout client.key \
  -out client.csr \
  -config client.cnf
```

## Sign CA

```bash
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key \
  -CAcreateserial -out server.crt -days 3650 -sha256
```

```bash
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key \
  -CAcreateserial -out client.crt -days 3650 -sha256
```
