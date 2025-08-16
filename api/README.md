# API Definitions

This directory contains API definitions for the Lumi-Go microservice.

## Structure

```
api/
├── openapi/        # OpenAPI/Swagger specifications
│   └── api.yaml   # Main OpenAPI spec
├── proto/         # Protocol Buffer definitions
│   └── service.proto  # gRPC service definitions
└── README.md
```

## OpenAPI (REST API)

The OpenAPI specification defines the RESTful HTTP API.

### Viewing the API Documentation

#### Option 1: Swagger UI (Online)
1. Visit [Swagger Editor](https://editor.swagger.io/)
2. Copy the contents of `openapi/api.yaml`
3. Paste into the editor

#### Option 2: Local Swagger UI
```bash
# Install swagger-ui
npm install -g @apidevtools/swagger-cli

# Serve the documentation
swagger-cli serve api/openapi/api.yaml
```

#### Option 3: ReDoc
```bash
# Install redoc-cli
npm install -g @redocly/cli

# Serve the documentation
redocly preview-docs api/openapi/api.yaml
```

### Generating Code from OpenAPI

#### Generate Server Stubs
```bash
# Install oapi-codegen
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# Generate server code
oapi-codegen -package api -generate types,server api/openapi/api.yaml > internal/api/openapi.gen.go
```

#### Generate Client SDK
```bash
# Generate Go client
oapi-codegen -package client -generate types,client api/openapi/api.yaml > client/openapi.gen.go

# Generate other language clients using OpenAPI Generator
docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
  -i /local/api/openapi/api.yaml \
  -g python \
  -o /local/client/python
```

### Validating OpenAPI Spec
```bash
# Using swagger-cli
swagger-cli validate api/openapi/api.yaml

# Using redocly
redocly lint api/openapi/api.yaml
```

## Protocol Buffers (gRPC API)

The Protocol Buffer definitions define the gRPC API.

### Prerequisites
```bash
# Install protoc
# macOS
brew install protobuf

# Linux
apt-get install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generating Code from Proto

```bash
# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/service.proto

# Generate with buf (recommended)
# Install buf
go install github.com/bufbuild/buf/cmd/buf@latest

# Generate code
buf generate
```

### buf.yaml Configuration
```yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

### buf.gen.yaml Configuration
```yaml
version: v1
plugins:
  - plugin: go
    out: gen/go
    opt: paths=source_relative
  - plugin: go-grpc
    out: gen/go
    opt: paths=source_relative
  - plugin: grpc-gateway
    out: gen/go
    opt: paths=source_relative
  - plugin: openapiv2
    out: gen/openapi
```

## Adding New Endpoints

### REST API (OpenAPI)

1. Edit `openapi/api.yaml`
2. Add your endpoint definition
3. Regenerate code if using code generation
4. Implement the handler in `internal/httpapi/`

Example:
```yaml
/api/v1/users/{id}:
  get:
    summary: Get user by ID
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    responses:
      '200':
        description: User found
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
```

### gRPC API (Proto)

1. Edit `proto/service.proto`
2. Add your service and message definitions
3. Generate code using protoc
4. Implement the service in `internal/rpcapi/`

Example:
```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

## API Versioning

### REST API Versioning
- Use URL path versioning: `/api/v1/`, `/api/v2/`
- Maintain backward compatibility within major versions
- Deprecate old versions with proper notice

### gRPC API Versioning
- Use package versioning: `lumigo.api.v1`, `lumigo.api.v2`
- Support multiple versions simultaneously
- Use field deprecation for gradual migration

## Best Practices

1. **Design First**: Define API before implementation
2. **Consistency**: Use consistent naming and patterns
3. **Documentation**: Keep API docs up-to-date
4. **Validation**: Validate requests and responses
5. **Error Handling**: Use standard error codes and formats
6. **Pagination**: Implement pagination for list endpoints
7. **Filtering**: Support filtering and sorting
8. **Rate Limiting**: Document rate limits
9. **Authentication**: Clearly define auth requirements
10. **Versioning**: Plan for API evolution

## Testing APIs

### Testing REST API
```bash
# Using curl
curl -X GET http://localhost:8080/api/v1/example

# Using httpie
http GET localhost:8080/api/v1/example

# Using Postman
# Import api/openapi/api.yaml into Postman
```

### Testing gRPC API
```bash
# Using grpcurl
grpcurl -plaintext localhost:8081 list
grpcurl -plaintext localhost:8081 describe lumigo.api.v1.ExampleService
grpcurl -plaintext -d '{"id": "123"}' localhost:8081 lumigo.api.v1.ExampleService/GetExample

# Using Evans (interactive gRPC client)
evans -p 8081 -r
```

## API Gateway Integration

For production deployments, consider using an API Gateway:

- **Kong**: API management and microservice management
- **Traefik**: Modern reverse proxy with auto-discovery
- **Istio**: Service mesh with advanced traffic management
- **AWS API Gateway**: Managed API gateway for AWS
- **Google Cloud Endpoints**: API management for GCP

## Monitoring

### Metrics to Track
- Request rate
- Response time (p50, p95, p99)
- Error rate
- Request size/response size
- Active connections

### Tools
- Prometheus + Grafana for metrics
- Jaeger/Zipkin for distributed tracing
- ELK Stack for log aggregation
