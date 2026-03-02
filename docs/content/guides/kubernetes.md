---
sidebar_position: 7
title: "Deploying to Kubernetes"
---

# Deploying to Kubernetes

The repository includes Kubernetes manifests in `deploy/k8s/` for deploying aether-webd to a cluster.

## Apply the manifests

```bash
kubectl apply -f deploy/k8s/
```

This creates a Deployment and a ClusterIP Service. The deployment runs a single replica with security-hardened defaults (non-root user, read-only root filesystem, dropped capabilities).

## Verify the deployment

Check that the pod is running:

```bash
kubectl get pods -l app=aether-webd
```

Check the service:

```bash
kubectl get svc aether-webd
```

The default service type is `ClusterIP` on port `8186`. To expose the service externally, edit `deploy/k8s/service.yaml` and uncomment the `LoadBalancer` or `NodePort` type.

## View logs

```bash
kubectl logs -l app=aether-webd
```

Follow logs in real time:

```bash
kubectl logs -f -l app=aether-webd
```

## Enable TLS

### Create a TLS secret

```bash
kubectl create secret tls aether-webd-tls \
  --cert=cert.pem \
  --key=key.pem
```

### Mount the secret in the deployment

Edit `deploy/k8s/deployment.yaml` and uncomment the volume and volume mount sections:

```yaml
containers:
  - name: aether-webd
    # ...
    args:
      - --listen
      - "0.0.0.0:8186"
      - --tls-cert
      - /etc/aether-webd/tls/tls.crt
      - --tls-key
      - /etc/aether-webd/tls/tls.key
    volumeMounts:
      - name: tls-certs
        mountPath: /etc/aether-webd/tls
        readOnly: true
volumes:
  - name: tls-certs
    secret:
      secretName: aether-webd-tls
```

Note: The commented-out volume mount sections already exist in the deployment manifest. Uncomment them and add the `--tls-cert` and `--tls-key` arguments to enable TLS.

Re-apply the updated manifest:

```bash
kubectl apply -f deploy/k8s/
```

## Set the API token

Store the token in a Kubernetes secret:

```bash
kubectl create secret generic aether-webd-token \
  --from-literal=api-token="$(openssl rand -hex 32)"
```

Reference the secret as an environment variable in the deployment:

```yaml
containers:
  - name: aether-webd
    env:
      - name: AETHER_API_TOKEN
        valueFrom:
          secretKeyRef:
            name: aether-webd-token
            key: api-token
```

## Resource limits

The default manifest includes resource requests and limits:

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 100m | 500m |
| Memory | 64Mi | 256Mi |

Adjust these in `deploy/k8s/deployment.yaml` based on your workload. Metric collection and task execution may require higher limits in production.

## Health checks

The deployment includes liveness and readiness probes using TCP socket checks on port `8186`. The probes confirm the server is accepting connections:

- **Liveness probe**: restarts the container if the server stops responding (initial delay: 5s, period: 10s).
- **Readiness probe**: removes the pod from the service endpoint pool until it is ready (initial delay: 5s, period: 5s).
