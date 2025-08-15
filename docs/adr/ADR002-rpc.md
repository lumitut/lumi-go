# ADR 002: RPC Framework

## Status
**Accepted** - August 2025

## Context

The lumi-go microservice template requires an RPC (Remote Procedure Call) framework for efficient service-to-service communication. While REST/HTTP APIs serve external clients well, internal service communication benefits from:

- Binary protocols with lower latency and bandwidth usage
- Strong typing with code generation
- Bidirectional streaming capabilities
- Native load balancing and retries
- Better performance for high-frequency calls

Key requirements for our RPC framework:
- Protocol buffer (protobuf) support for schema definition
- HTTP/2 transport for multiplexing and flow control
- Compatibility with both gRPC and REST clients
- Browser/web client support without proxies
- Strong ecosystem and tooling
- OpenTelemetry integration for tracing
- Graceful degradation and backward compatibility
- Developer-friendly debugging tools

The RPC framework must coexist with our HTTP framework (Gin) and support gradual migration from REST to RPC where beneficial.

## Decision

We have chosen **Connect** (connectrpc.com/connect) as our RPC framework, built on top of gRPC protocols with enhanced flexibility.

### Key factors in this decision:

1. **Protocol Flexibility**: Connect supports three protocols from the same service:
   - gRPC - for service-to-service communication
   - gRPC-Web - for browser clients
   - Connect's own protocol - JSON over HTTP/1.1 for compatibility

2. **Developer Experience**:
   - Works with standard `net/http` handlers
   - Can be mounted alongside Gin routes
   - Human-readable JSON for Connect protocol
   - No proxy required for web clients
   - curl-able endpoints for debugging

3. **Code Generation**:
   - Built on standard protobuf
   - Type-safe client and server code
   - Supports proto3 and proto2
   - Compatible with buf toolchain

4. **Production Features**:
   - Built-in compression (gzip, br)
   - Interceptors for cross-cutting concerns
   - Streaming support (client, server, bidirectional)
   - Context propagation and cancellation

5. **Compatibility**:
   - Full gRPC compatibility
   - Works with existing gRPC clients
   - Supports gRPC ecosystem tools
   - Easy migration path from/to pure gRPC

### Implementation Architecture:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │     │  Mobile App │     │  Service B  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                    │
   Connect/JSON        gRPC-Web              gRPC
       │                   │                    │
       └───────────────────┼────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   Connect   │
                    │   Handler   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Service   │
                    │    Logic    │
                    └─────────────┘
```

## Consequences

### Positive Consequences

- **Universal Clients**: Single service works with browsers, mobile, and services
- **No Proxy Required**: Direct browser-to-service communication
- **Debugging Friendly**: JSON protocol can be debugged with curl/Postman
- **Gradual Migration**: Can run alongside REST APIs
- **Standard Compliance**: Full gRPC compatibility maintained
- **Type Safety**: Generated code prevents runtime errors
- **Performance**: Binary protocol for service-to-service calls

### Negative Consequences

- **Less Mature**: Newer than pure gRPC (Connect released 2022)
- **Smaller Community**: Fewer resources than gRPC
- **Additional Complexity**: Three protocols to understand
- **Limited Language Support**: Primarily Go and TypeScript/JavaScript
- **Buf Dependency**: Best experience requires buf toolchain

### Mitigations

- Maintain gRPC compatibility for easy migration if needed
- Document protocol selection guidelines
- Provide debugging tools and examples
- Use Connect protocol only where JSON is beneficial
- Establish clear service boundaries to minimize RPC surface

## Alternatives Considered

### Option 1: gRPC (google.golang.org/grpc)

**Pros:**
- Industry standard with massive adoption
- Support for 11+ languages
- Battle-tested at scale
- Extensive tooling ecosystem
- Native Kubernetes support

**Cons:**
- Requires grpc-web proxy for browsers
- Binary protocol hard to debug
- No native JSON support
- Steeper learning curve
- HTTP/2 only (problematic for some proxies)

**Reason not chosen:** Lack of browser support without proxy and poor debugging experience.

### Option 2: Twirp (github.com/twitchtv/twirp)

**Pros:**
- Simple and pragmatic
- JSON and protobuf support
- Works over HTTP/1.1
- Easy to debug
- Small and focused

**Cons:**
- Limited features (no streaming)
- Smaller community
- No gRPC compatibility
- Less active development
- Fewer language implementations

**Reason not chosen:** Lack of streaming support and limited feature set.

### Option 3: JSON-RPC 2.0

**Pros:**
- Simple protocol
- Human-readable
- Wide language support
- HTTP/1.1 compatible
- Easy to implement

**Cons:**
- No code generation
- No type safety
- No streaming support
- No binary protocol option
- Manual schema management

**Reason not chosen:** Lack of type safety and code generation.

### Option 4: GraphQL

**Pros:**
- Flexible query language
- Single endpoint
- Strong typing
- Great for frontend needs
- Rich ecosystem

**Cons:**
- Complexity overhead
- N+1 query problems
- Caching challenges
- Not ideal for service-to-service
- No streaming support

**Reason not chosen:** Better suited for frontend APIs than service-to-service communication.

### Option 5: REST/HTTP Only

**Pros:**
- Universal support
- Simple to understand
- Great tooling
- Human-readable
- Cache-friendly

**Cons:**
- No code generation
- Higher latency
- More bandwidth usage
- No streaming support
- Manual contract management

**Reason not chosen:** Inefficient for high-frequency internal service calls.

## Implementation Guidelines

### Service Design Principles

1. **Use RPC for**:
   - Internal service-to-service calls
   - High-frequency, low-latency operations
   - Strongly-typed data exchange
   - Streaming data scenarios

2. **Use REST for**:
   - Public APIs
   - CRUD operations
   - Cache-heavy endpoints
   - Third-party integrations

### Protocol Selection

```
if browser_client && no_streaming:
    use Connect (JSON)
elif browser_client && streaming:
    use gRPC-Web
elif service_to_service:
    use gRPC
elif debugging:
    use Connect (JSON)
```

### Proto Style Guide

```protobuf
syntax = "proto3";

package lumitut.service.v1;

import "google/protobuf/timestamp.proto";
import "validate/validate.proto";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc WatchUsers(WatchUsersRequest) returns (stream User);
}

message GetUserRequest {
  string id = 1 [(validate.rules).string.uuid = true];
}

message GetUserResponse {
  User user = 1;
}

message User {
  string id = 1;
  string email = 2;
  google.protobuf.Timestamp created_at = 3;
}
```

### Interceptor Stack

1. Recovery (panic handling)
2. Request ID propagation
3. Authentication/Authorization
4. Rate limiting
5. OpenTelemetry tracing
6. Structured logging
7. Metrics collection
8. Retry logic (client-side)

## Migration Strategy

1. **Phase 1**: Set up Connect alongside Gin
2. **Phase 2**: Migrate internal service calls to RPC
3. **Phase 3**: Create gRPC clients for service communication
4. **Phase 4**: Expose Connect endpoints for web clients
5. **Phase 5**: Deprecate redundant REST endpoints

## Related

- ADR 001: Go Web Framework
- ADR 003: API Design Patterns
- ADR 004: Observability Stack
- [Connect Documentation](https://connect.build/docs/go/)
- [Protocol Buffer Style Guide](https://developers.google.com/protocol-buffers/docs/style)
- [gRPC Best Practices](https://grpc.io/docs/guides/)
