---
sidebar_position: 5
title: "Security"
---

# Security

This guide covers enabling TLS, mutual TLS (mTLS), and token authentication for aether-webd.

## Enable TLS with auto-generated certificates

```bash
aether-webd --tls
```

Or via the systemd environment file (`/etc/aether-webd/env`):

```
AETHER_TLS=true
```

On first run, this generates a local CA and CA-signed server certificate under `{data-dir}/certs/` (default: `/var/lib/aether-webd/certs/`). The generated files are reused on subsequent starts.

| File | Description |
|------|-------------|
| `ca.pem` | CA certificate -- add to browser/system trust store |
| `ca-key.pem` | CA private key |
| `server.pem` | Server certificate signed by the CA |
| `server-key.pem` | Server private key |

### Trust the CA

To suppress browser warnings, add the CA to your system trust store:

```bash
sudo cp /var/lib/aether-webd/certs/ca.pem /usr/local/share/ca-certificates/aether-webd.crt
sudo update-ca-certificates
```

For curl, pass the CA explicitly:

```bash
curl --cacert /var/lib/aether-webd/certs/ca.pem https://localhost:8186/api/v1/meta/version
```

To regenerate certificates (e.g., after expiry), delete the `certs/` directory and restart aether-webd.

## Enable TLS with your own certificates

```bash
aether-webd --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem
```

Both flags are required together. Providing only one is an error.

## Enable mutual TLS (mTLS)

mTLS requires clients to present a certificate signed by a trusted CA. This is well-suited for machine-to-machine communication.

```bash
aether-webd --tls --mtls-ca-cert /path/to/client-ca.pem
```

The `--mtls-ca-cert` flag implicitly enables TLS. If no `--tls-cert`/`--tls-key` pair is provided, the server auto-generates its own server certificate.

### Create a CA and client certificate

```bash
# Generate CA key and certificate
openssl ecparam -genkey -name prime256v1 -out ca-key.pem
openssl req -new -x509 -key ca-key.pem -out ca.pem -days 365 -subj "/CN=Aether CA"

# Generate client key and CSR
openssl ecparam -genkey -name prime256v1 -out client-key.pem
openssl req -new -key client-key.pem -out client.csr -subj "/CN=client"

# Sign client certificate with CA
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out client.pem -days 365
```

### Connect with a client certificate

```bash
curl --cacert /var/lib/aether-webd/certs/ca.pem \
     --cert client.pem \
     --key client-key.pem \
     https://localhost:8186/api/v1/meta/version
```

Clients without a valid certificate are rejected at the TLS handshake layer before any HTTP response is sent.

## Enable token authentication

Token authentication protects all `/api/*` endpoints with a bearer token.

### Set the token

Via CLI flag:

```bash
aether-webd --api-token mysecrettoken
```

Via environment variable (the flag takes precedence if both are set):

```bash
export AETHER_API_TOKEN=mysecrettoken
aether-webd
```

Via the systemd environment file (`/etc/aether-webd/env`):

```
AETHER_API_TOKEN=mysecrettoken
```

### Make authenticated requests

```bash
curl -H "Authorization: Bearer mysecrettoken" http://localhost:8186/api/v1/meta/version
```

### Exempt paths

These paths do not require a token:

| Path | Reason |
|------|--------|
| `/healthz` | Health checks and load balancer probes |
| `/openapi.json` | API schema discovery |
| `/docs`, `/docs/*` | Built-in API documentation |
| Paths not under `/api/` | Frontend static files |

### Error response

Unauthorized requests receive a `401` JSON response:

```json
{
  "status": 401,
  "title": "Unauthorized",
  "detail": "missing Authorization header"
}
```

## Activation matrix

| Flags | Result |
|-------|--------|
| _(none)_ | HTTP, no TLS |
| `--tls` | HTTPS, auto-generated self-signed cert |
| `--tls-cert` + `--tls-key` | HTTPS, user-provided cert |
| `--tls-cert` only | Error: both `--tls-cert` and `--tls-key` required |
| `--mtls-ca-cert` | HTTPS + mTLS, auto-generated server cert |
| `--mtls-ca-cert` + `--tls-cert` + `--tls-key` | HTTPS + mTLS, user-provided server cert |
| `--tls` + `--tls-cert` + `--tls-key` | HTTPS, user-provided cert (`--tls` redundant) |
| `--api-token` | Token auth on `/api/*` paths (combinable with any TLS mode) |

## Production recommendations

- Use certificates from a trusted CA (e.g., Let's Encrypt) rather than the auto-generated self-signed certs.
- Enable mTLS for machine-to-machine communication between services.
- Generate strong, random API tokens: `openssl rand -hex 32`.
- Set the token via the `AETHER_API_TOKEN` environment variable (or in `/etc/aether-webd/env` for systemd) to keep it out of process listings.
- Place aether-webd behind a reverse proxy (e.g., Caddy, nginx) for TLS termination in production.
- Minimum TLS version is 1.2; TLS 1.0 and 1.1 are not supported.

## Combined example: TLS + mTLS + token

```bash
aether-webd \
  --tls-cert /etc/aether/server.pem \
  --tls-key /etc/aether/server-key.pem \
  --mtls-ca-cert /etc/aether/ca.pem \
  --api-token "$AETHER_API_TOKEN"
```

```bash
curl --cacert /etc/aether/ca.pem \
     --cert /etc/aether/client.pem \
     --key /etc/aether/client-key.pem \
     -H "Authorization: Bearer $AETHER_API_TOKEN" \
     https://aether.example.com:8186/api/v1/onramp/state
```
