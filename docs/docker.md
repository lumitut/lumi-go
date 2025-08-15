# Docker Configuration for lumi-go

## Overview

This directory contains Docker configurations for building and running the Go Middle-Service Template (GMT) application.

## Files

- `Dockerfile` - Production multi-stage build
- `Dockerfile.dev` - Development image with hot-reload
- `build.sh` - Build script with security scanning
- `.dockerignore` - Files to exclude from build context

## Building Images

### Production Build

```bash
# Simple build
./deploy/docker/build.sh

# Build and push to registry
./deploy/docker/build.sh --push

# Build with custom version
./deploy/docker/build.sh --version v1.2.3

# Build for specific platform
./deploy/docker/build.sh --platform linux/amd64

# Skip security scanning
./deploy/docker/build.sh --no-scan
```

### Development Build

```bash
# Build development image
./deploy/docker/build.sh --dev

# Or directly with Docker
docker build -f deploy/docker/Dockerfile.dev -t lumi-go:dev .
```

### Manual Docker Commands

```bash
# Build production image
docker build -f deploy/docker/Dockerfile -t lumi-go:latest .

# Build with build args
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -f deploy/docker/Dockerfile \
  -t lumi-go:v1.0.0 .

# Multi-platform build (requires buildx)
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f deploy/docker/Dockerfile \
  -t lumi-go:latest \
  --push .
```

## Running Containers

### Production Container

```bash
# Run with defaults
docker run -p 8080:8080 lumi-go:latest

# Run with environment variables
docker run \
  -p 8080:8080 \
  -e LOG_LEVEL=debug \
  -e PG_URL=postgres://user:pass@host:5432/db \
  lumi-go:latest

# Run with health check
docker run \
  -p 8080:8080 \
  --health-cmd="/service health" \
  --health-interval=30s \
  lumi-go:latest
```

### Development Container

```bash
# Run with hot-reload
docker run \
  -v $(pwd):/app \
  -p 8080:8080 \
  -p 2345:2345 \
  lumi-go:dev

# Run with debugger
docker run \
  -v $(pwd):/app \
  -p 8080:8080 \
  -p 2345:2345 \
  --security-opt="apparmor=unconfined" \
  --cap-add=SYS_PTRACE \
  lumi-go:dev dlv debug --headless --listen=:2345 --api-version=2
```

## Security Features

### Production Image

- **Distroless base**: Minimal attack surface
- **Non-root user**: Runs as UID 65532
- **No shell**: Cannot exec into container
- **Static binary**: No dynamic dependencies
- **Security scanning**: Trivy scan during build

### Build-time Security

- **Multi-stage build**: Source code not in final image
- **Minimal layers**: Reduced attack surface
- **Version pinning**: Reproducible builds
- **Build args**: No secrets in image layers

### Runtime Security

```bash
# Run with security options
docker run \
  --read-only \
  --security-opt=no-new-privileges \
  --cap-drop=ALL \
  -p 8080:8080 \
  lumi-go:latest
```

## Image Optimization

### Size Comparison

| Build Type | Base Image | Size |
|------------|------------|------|
| Development | golang:1.22-alpine | ~500MB |
| Production (alpine) | alpine:3.19 | ~20MB |
| Production (distroless) | distroless/static | ~15MB |

### Optimization Techniques

1. **Multi-stage builds**: Separate build and runtime
2. **Static linking**: No dynamic libraries
3. **Binary stripping**: Remove debug symbols
4. **Distroless base**: Minimal runtime
5. **Layer caching**: Optimize build speed

## Container Registry

### Docker Hub

```bash
# Login to Docker Hub
docker login

# Tag for Docker Hub
docker tag lumi-go:latest lumitut/lumi-go:latest

# Push to Docker Hub
docker push lumitut/lumi-go:latest
```

### AWS ECR

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Tag for ECR
docker tag lumi-go:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/lumi-go:latest

# Push to ECR
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/lumi-go:latest
```

### GitHub Container Registry

```bash
# Login to ghcr.io
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Tag for GHCR
docker tag lumi-go:latest ghcr.io/lumitut/lumi-go:latest

# Push to GHCR
docker push ghcr.io/lumitut/lumi-go:latest
```

## Docker Compose

For local development with dependencies, use docker-compose:

```bash
# Start all services
docker-compose up

# Start in background
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## Troubleshooting

### Build Issues

```bash
# Clear Docker cache
docker builder prune -a

# Build with no cache
docker build --no-cache -f deploy/docker/Dockerfile -t lumi-go:latest .

# Inspect build context size
du -sh .

# List build context files
tar -czf - . | tar -tzf - | head -20
```

### Runtime Issues

```bash
# Check container logs
docker logs <container-id>

# Inspect container
docker inspect <container-id>

# Check resource usage
docker stats <container-id>

# For distroless debugging, use debug image
docker run --rm -it --entrypoint=sh gcr.io/distroless/static:debug
```

### Security Scanning

```bash
# Scan with Trivy
trivy image lumi-go:latest

# Scan with Docker Scout
docker scout cves lumi-go:latest

# Scan with Snyk
snyk container test lumi-go:latest
```

## Best Practices

1. **Never include secrets in images**
2. **Use specific version tags, not latest**
3. **Scan images before deployment**
4. **Sign images for production**
5. **Use minimal base images**
6. **Run as non-root user**
7. **Set resource limits**
8. **Use health checks**
9. **Label images properly**
10. **Clean up unused images regularly**

## CI/CD Integration

See `.github/workflows/` for GitHub Actions workflows that:
- Build and test images on PR
- Scan for vulnerabilities
- Push to registry on merge
- Deploy to Kubernetes

## Additional Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Container Security Guide](https://sysdig.com/learn-cloud-native/kubernetes-security/container-security-best-practices/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Docker BuildKit](https://docs.docker.com/develop/develop-images/build_enhancements/)
