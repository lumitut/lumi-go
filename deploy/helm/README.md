# lumi-go Helm Chart

## Overview

This Helm chart deploys the Go Middle-Service Template (GMT) application on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.10+
- PV provisioner support in the underlying infrastructure (if persistence is required)

## Installation

### Add the repository (when published)

```bash
helm repo add lumitut https://charts.lumitut.com
helm repo update
```

### Install the chart

```bash
# Install with default values
helm install my-service lumitut/lumi-go

# Install with custom values
helm install my-service lumitut/lumi-go -f values.yaml

# Install in a specific namespace
helm install my-service lumitut/lumi-go --namespace my-namespace --create-namespace
```

### Install from source

```bash
# From the repository root
helm install my-service ./deploy/helm

# With custom values
helm install my-service ./deploy/helm -f my-values.yaml
```

## Upgrading

```bash
# Upgrade to a new version
helm upgrade my-service lumitut/lumi-go

# Upgrade with new values
helm upgrade my-service lumitut/lumi-go -f values.yaml

# Rollback if needed
helm rollback my-service
```

## Uninstallation

```bash
helm uninstall my-service
```

## Configuration

See [values.yaml](values.yaml) for the full list of configurable parameters.

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `image.repository` | Container image repository | `lumitut/lumi-go` |
| `image.tag` | Container image tag | `""` (uses chart appVersion) |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `ingress.enabled` | Enable ingress | `false` |
| `resources` | CPU/Memory resource requests/limits | See values.yaml |
| `autoscaling.enabled` | Enable HPA | `true` |
| `database.enabled` | Enable database configuration | `false` |
| `redis.enabled` | Enable Redis configuration | `false` |

### Environment-specific Values

Create environment-specific values files:

```yaml
# values-dev.yaml
config:
  env: dev
  logLevel: debug
  ginMode: debug

replicaCount: 1

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

```yaml
# values-prod.yaml
config:
  env: prod
  logLevel: info
  ginMode: release

replicaCount: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

## Security Considerations

### Pod Security

The chart implements several security best practices:

- Runs as non-root user (UID 65534)
- Read-only root filesystem
- Drops all capabilities
- No privilege escalation

### Network Policies

Enable network policies to restrict traffic:

```yaml
networkPolicy:
  enabled: true
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: database
```

### Secrets Management

Use existing secrets for sensitive data:

```yaml
database:
  existingSecret: "my-db-secret"
  passwordKey: "password"

redis:
  existingSecret: "my-redis-secret"
  passwordKey: "password"

featureFlags:
  existingSecret: "my-ff-secret"
  tokenKey: "token"
```

## Monitoring

### Prometheus Metrics

Enable ServiceMonitor for Prometheus Operator:

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  labels:
    prometheus: kube-prometheus
```

### Health Checks

The chart configures liveness and readiness probes:

- Liveness: `/healthz` endpoint
- Readiness: `/readyz` endpoint

## Troubleshooting

### Debug deployment issues

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=lumi-go

# View pod logs
kubectl logs -l app.kubernetes.io/name=lumi-go

# Describe pod for events
kubectl describe pod <pod-name>

# Check service endpoints
kubectl get endpoints <service-name>
```

### Common Issues

1. **Pods not starting**: Check resource limits and node capacity
2. **Readiness probe failing**: Verify database/Redis connectivity
3. **High memory usage**: Adjust resource limits and check for memory leaks
4. **Ingress not working**: Verify ingress controller and DNS configuration

## Development

### Testing the Chart

```bash
# Lint the chart
helm lint ./deploy/helm

# Dry run installation
helm install my-service ./deploy/helm --dry-run --debug

# Template rendering
helm template my-service ./deploy/helm

# Test with different values
helm template my-service ./deploy/helm -f values-dev.yaml
```

### Chart Testing with ct

```bash
# Install chart-testing tool
brew install chart-testing

# Lint chart
ct lint --charts ./deploy/helm

# Install and test
ct install --charts ./deploy/helm
```

## Contributing

Please see [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines on contributing to this chart.

## License

This chart is licensed under the MIT License. See [LICENSE](../../LICENSE) for details.
