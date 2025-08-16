# External Services Integration

The `lumi-go` template is designed as a lean microservice that can optionally connect to external services. This document explains how to integrate with databases, caches, and observability platforms.

## Architecture Philosophy

This template follows a **microservice-first** approach:
- The core service is self-contained and can run independently
- External dependencies are optional and configured via clients
- Infrastructure services are managed separately
- Cloud-managed services are preferred for production

## Configuration Pattern

All external services follow the same configuration pattern:

```json
{
  "clients": {
    "serviceName": {
      "enabled": false,
      "url": "connection-string"
    }
  }
}
```

Environment variables:
```bash
LUMI_CLIENTS_SERVICENAME_ENABLED=true
LUMI_CLIENTS_SERVICENAME_URL=connection-string
```

## Database Integration

### PostgreSQL

For local development, use the `lumi-postgres` template:
```bash
# Start PostgreSQL
cd ../lumi-postgres
docker-compose up -d

# Configure lumi-go
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

For production, use managed services:
- **AWS**: RDS PostgreSQL
- **Azure**: Database for PostgreSQL
- **GCP**: Cloud SQL for PostgreSQL

### Connection String Format
```
postgres://username:password@host:port/database?sslmode=disable
```

### Using in Code
```go
func NewDatabaseClient(cfg *config.Config) (*sql.DB, error) {
    dbURL, enabled := cfg.GetDatabaseURL()
    if !enabled {
        return nil, nil // Database not configured
    }
    
    return sql.Open("postgres", dbURL)
}
```

## Redis Integration

### Local Development

Use the `lumi-redis` template:
```bash
# Start Redis
cd ../lumi-redis
docker-compose up -d

# Configure lumi-go
export LUMI_CLIENTS_REDIS_ENABLED=true
export LUMI_CLIENTS_REDIS_URL=redis://localhost:6379/0
```

### Production Options
- **AWS**: ElastiCache
- **Azure**: Cache for Redis
- **GCP**: Memorystore for Redis

### Connection String Format
```
redis://[:password@]host[:port][/database]
```

### Using in Code
```go
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
    redisURL, enabled := cfg.GetRedisURL()
    if !enabled {
        return nil, nil // Redis not configured
    }
    
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }
    
    return redis.NewClient(opt), nil
}
```

## Observability

### Metrics (Prometheus)

The service exposes metrics at `/metrics` endpoint. No configuration needed for basic metrics.

For metric collection:
```bash
# Local Prometheus
docker run -p 9090:9090 prom/prometheus \
  --config.file=/etc/prometheus/prometheus.yml

# Or use managed services:
# - AWS CloudWatch
# - Azure Monitor
# - GCP Cloud Monitoring
# - Grafana Cloud
```

### Distributed Tracing

Configure OpenTelemetry endpoint:
```bash
export LUMI_CLIENTS_TRACING_ENABLED=true
export LUMI_CLIENTS_TRACING_ENDPOINT=localhost:4317
```

For production:
- **AWS**: X-Ray
- **Azure**: Application Insights
- **GCP**: Cloud Trace
- **Datadog APM**
- **New Relic**

### Logging

Logs are written to stdout in JSON format by default. Configure log aggregation:

Local:
```bash
# View logs
docker logs lumi-go-app

# Or use log aggregation
docker run -p 5601:5601 -p 9200:9200 -p 5044:5044 \
  sebp/elk
```

Production:
- **AWS**: CloudWatch Logs
- **Azure**: Log Analytics
- **GCP**: Cloud Logging
- **ELK Stack**
- **Splunk**

## Message Queues

While not included in the base template, you can add message queue clients:

### Kafka Example
```go
// Add to config
type KafkaClientConfig struct {
    Enabled bool     `json:"enabled"`
    Brokers []string `json:"brokers"`
}

// Usage
if cfg.Clients.Kafka.Enabled {
    producer, err := kafka.NewProducer(cfg.Clients.Kafka.Brokers)
    // ...
}
```

### Cloud Options
- **AWS**: SQS, SNS, Kinesis
- **Azure**: Service Bus, Event Hubs
- **GCP**: Pub/Sub

## Development Workflow

### 1. Start with Core Service
```bash
# Run just the microservice
make run
```

### 2. Add Services as Needed
```bash
# Need a database?
cd ../lumi-postgres && docker-compose up -d
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://localhost:5432/mydb

# Need caching?
cd ../lumi-redis && docker-compose up -d
export LUMI_CLIENTS_REDIS_ENABLED=true
export LUMI_CLIENTS_REDIS_URL=redis://localhost:6379

# Restart service
make run
```

### 3. Use docker-compose.dev.yml for Development
```yaml
version: '3.8'
services:
  app:
    # ... app config ...
    environment:
      - LUMI_CLIENTS_DATABASE_ENABLED=true
      - LUMI_CLIENTS_DATABASE_URL=postgres://host.docker.internal:5432/db
      - LUMI_CLIENTS_REDIS_ENABLED=true
      - LUMI_CLIENTS_REDIS_URL=redis://host.docker.internal:6379
```

## Production Deployment

### 1. Use Environment Variables
```yaml
# Kubernetes ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: lumi-go-config
data:
  LUMI_SERVICE_ENVIRONMENT: production
  LUMI_CLIENTS_DATABASE_ENABLED: "true"
```

### 2. Use Secrets for Sensitive Data
```yaml
# Kubernetes Secret
apiVersion: v1
kind: Secret
metadata:
  name: lumi-go-secrets
stringData:
  LUMI_CLIENTS_DATABASE_URL: postgres://user:pass@rds.amazonaws.com/db
```

### 3. Service Discovery

For Kubernetes:
```bash
# Database service
LUMI_CLIENTS_DATABASE_URL=postgres://postgres-service:5432/db

# Redis service
LUMI_CLIENTS_REDIS_URL=redis://redis-service:6379
```

## Best Practices

1. **Start Simple**: Begin with just the microservice, add services as needed
2. **Use Managed Services**: Prefer cloud-managed services for production
3. **Environment-Specific Config**: Use different configs for dev/staging/prod
4. **Connection Pooling**: Implement proper connection pooling for databases
5. **Circuit Breakers**: Add circuit breakers for external service calls
6. **Health Checks**: Include external service checks in readiness probes
7. **Graceful Degradation**: Service should function with degraded features if external services fail

## Example: Full Stack Local Development

```bash
# 1. Start infrastructure services
cd ../lumi-postgres && docker-compose up -d
cd ../lumi-redis && docker-compose up -d
cd ../lumi-grafana && docker-compose up -d

# 2. Configure and run lumi-go
cd ../lumi-go
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://lumigo:lumigo@localhost:5432/lumigo
export LUMI_CLIENTS_REDIS_ENABLED=true
export LUMI_CLIENTS_REDIS_URL=redis://localhost:6379
export LUMI_CLIENTS_TRACING_ENABLED=true
export LUMI_CLIENTS_TRACING_ENDPOINT=localhost:4317

# 3. Run with hot reload
make run-dev

# 4. Access services
# - API: http://localhost:8080
# - Metrics: http://localhost:9090/metrics
# - Grafana: http://localhost:3000
```

## Troubleshooting

### Connection Refused
- Check if external service is running
- Verify port numbers and host names
- Use `host.docker.internal` when connecting from Docker to host services

### Authentication Failed
- Verify credentials in connection string
- Check if user has proper permissions
- Ensure SSL mode matches server configuration

### Performance Issues
- Implement connection pooling
- Add caching layer
- Use read replicas for databases
- Consider async processing for heavy operations
