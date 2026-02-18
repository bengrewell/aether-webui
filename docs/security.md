# Security Configuration

## TLS

TLS encrypts traffic between clients and the server.

### Auto-Generated Certificate

```bash
aether-webd --tls
```

On first run, generates a local CA and CA-signed server certificate under `<data-dir>/certs/` (default: `/var/lib/aether-webd/certs/`). Certificates are persisted and reused on subsequent starts.

Generated files:

| File | Description |
|------|-------------|
| `ca.pem` | CA certificate — add this to your browser/system trust store |
| `ca-key.pem` | CA private key (owner-read only) |
| `server.pem` | Server certificate signed by the CA |
| `server-key.pem` | Server private key (owner-read only) |

Certificate properties:
- ECDSA P-256 keys, 1-year validity
- Server SANs: `localhost`, `127.0.0.1`, `::1`
- Server CN: `aether-webd`, CA CN: `Aether WebUI CA`

#### Trusting the CA

To avoid browser warnings, add the generated CA to your trust store:

**Firefox:**
1. Open Settings > Privacy & Security > Certificates > View Certificates
2. Import `<data-dir>/certs/ca.pem` under the Authorities tab
3. Check "Trust this CA to identify websites"

**Chrome / system (Debian/Ubuntu):**
```bash
sudo cp /var/lib/aether-webd/certs/ca.pem /usr/local/share/ca-certificates/aether-webd.crt
sudo update-ca-certificates
```

**curl:**
```bash
curl --cacert /var/lib/aether-webd/certs/ca.pem https://localhost:8186/api/v1/meta/version
```

To regenerate certificates (e.g., after expiry), delete the `certs/` directory and restart.

### User-Provided Certificate

```bash
aether-webd --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem
```

Both `--tls-cert` and `--tls-key` are required together. Providing only one is an error.

### Activation Matrix

| Flags | Result |
|-------|--------|
| _(none)_ | HTTP, no TLS |
| `--tls` | HTTPS, auto-generated self-signed cert |
| `--tls-cert` + `--tls-key` | HTTPS, user-provided cert |
| `--tls-cert` only | Error: both required |
| `--mtls-ca-cert` | HTTPS + mTLS, auto-generated server cert |
| `--mtls-ca-cert` + `--tls-cert` + `--tls-key` | HTTPS + mTLS, user-provided cert |
| `--tls` + `--tls-cert` + `--tls-key` | HTTPS, user-provided cert (`--tls` redundant) |

## Mutual TLS (mTLS)

mTLS requires clients to present a certificate signed by a trusted CA.

```bash
aether-webd --tls --mtls-ca-cert /path/to/ca.pem
```

### Creating a CA and Client Certificate

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

# Connect with client certificate
curl -k --cert client.pem --key client-key.pem https://localhost:8186/api/v1/meta/version
```

Clients without a valid certificate receive a TLS handshake error (connection refused at the TLS layer, before any HTTP response).

## Token Authentication

Bearer token authentication protects API endpoints.

### Configuration

Via CLI flag:
```bash
aether-webd --api-token mysecrettoken
```

Via environment variable (flag takes precedence):
```bash
export AETHER_API_TOKEN=mysecrettoken
aether-webd
```

### Request Format

```bash
curl -H "Authorization: Bearer mysecrettoken" http://localhost:8186/api/v1/meta/version
```

The `Bearer` prefix is case-insensitive. Token comparison uses constant-time comparison to prevent timing attacks.

### Exempt Paths

The following paths do **not** require a token:

| Path | Reason |
|------|--------|
| `/healthz` | Health checks / load balancer probes |
| `/openapi.json` | API schema discovery |
| `/docs`, `/docs/*` | API documentation |
| Paths not under `/api/` | Frontend static files |

All `/api/*` paths require a valid token when authentication is enabled.

### Error Response

Unauthorized requests receive a `401` JSON response:

```json
{
  "status": 401,
  "title": "Unauthorized",
  "detail": "missing Authorization header"
}
```

## Combined Usage

### Development (TLS + token)

```bash
aether-webd --tls --api-token dev-token-123
```

### Production (user certs + mTLS + token)

```bash
aether-webd \
  --tls-cert /etc/aether/server.pem \
  --tls-key /etc/aether/server-key.pem \
  --mtls-ca-cert /etc/aether/ca.pem \
  --api-token "$AETHER_API_TOKEN"
```

### Production Recommendations

- Use certificates from a trusted CA (e.g., Let's Encrypt) rather than self-signed certs
- Enable mTLS for machine-to-machine communication
- Use strong, randomly generated API tokens (e.g., `openssl rand -hex 32`)
- Set the token via `AETHER_API_TOKEN` env var to avoid exposing it in process listings
- Place the server behind a reverse proxy (e.g., Caddy, nginx) for TLS termination in production
- Minimum TLS version is 1.2; TLS 1.0 and 1.1 are not supported
