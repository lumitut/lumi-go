# ADR 001: Go Web Framework

## Status
**Accepted** - August 2025

## Context

The lumi-go microservice template requires a robust HTTP web framework to handle REST API endpoints, middleware management, routing, and request/response processing. The framework choice significantly impacts:

- Developer productivity and experience
- Application performance and latency
- Middleware ecosystem availability
- Community support and documentation
- Long-term maintenance burden
- Integration with observability tools (OpenTelemetry, Prometheus)

Key requirements for our framework:
- High performance with low memory footprint
- Extensive middleware support
- Strong community and ecosystem
- Good documentation and learning resources
- Native support for JSON binding/validation
- Compatible with standard `net/http` handlers
- Production-proven at scale
- Active maintenance and security updates

## Decision

We have chosen **Gin** (github.com/gin-gonic/gin) as our HTTP web framework for the lumi-go template.

### Key factors in this decision:

1. **Performance**: Gin consistently ranks among the fastest Go web frameworks in benchmarks, using a custom version of HttpRouter with zero memory allocation for routing.

2. **Developer Experience**: 
   - Intuitive API design with minimal boilerplate
   - Excellent error handling and recovery mechanisms
   - Built-in validation using go-playground/validator
   - Comprehensive request binding (JSON, XML, YAML, form data)

3. **Ecosystem**: 
   - Largest collection of community middleware
   - Native integration with popular tools (Swagger, OpenTelemetry, Prometheus)
   - Extensive third-party library support

4. **Production Readiness**:
   - Battle-tested in production by major companies
   - Mature codebase with predictable behavior
   - Excellent debugging support with detailed panic recovery

5. **Documentation**: 
   - Comprehensive official documentation
   - Abundant tutorials and examples
   - Large community for support

## Consequences

### Positive Consequences

- **Fast Development**: Developers can quickly build features with Gin's intuitive API
- **Performance**: Low latency and high throughput for API endpoints
- **Middleware Reuse**: Access to extensive middleware ecosystem
- **Easy Testing**: Built-in testing utilities and httptest compatibility
- **Monitoring**: Native Prometheus metrics and OpenTelemetry integration
- **Team Scalability**: Easy to onboard new developers due to popularity

### Negative Consequences

- **Framework Lock-in**: Code becomes tightly coupled to Gin's API
- **Context Handling**: Gin uses custom context which can complicate some patterns
- **Breaking Changes**: Major version updates may require code refactoring
- **Overhead for Simple Services**: Might be overkill for very simple HTTP services

### Mitigations

- Wrap Gin-specific code in interfaces where possible
- Use standard `http.Handler` for framework-agnostic components
- Pin framework version and upgrade carefully with testing
- Consider using standard library for simple internal services

## Alternatives Considered

### Chi (github.com/go-chi/chi)

**Pros:**
- Lightweight and modular
- 100% compatible with net/http
- No external dependencies
- More idiomatic Go code
- Composable middleware stack

**Cons:**
- Less built-in functionality (requires more manual work)
- Smaller ecosystem of ready-made middleware
- No built-in validation or binding
- Performance slightly lower than Gin in benchmarks

**Reason not chosen:** While Chi's philosophy of staying close to standard library is admirable, Gin's batteries-included approach and superior performance better suit our need for rapid development and high throughput.

### Echo (github.com/labstack/echo)

**Pros:**
- High performance
- Good middleware collection
- Built-in JWT support
- Automatic TLS certificates

**Cons:**
- Smaller community than Gin
- Less mature ecosystem
- Documentation not as comprehensive
- Custom HTTP error handling can be complex

**Reason not chosen:** Echo is excellent but has a smaller community and ecosystem compared to Gin, which could impact long-term support and middleware availability.

### Fiber (github.com/gofiber/fiber)

**Pros:**
- Fastest performance (built on Fasthttp)
- Express-like API (familiar to Node.js developers)
- Built-in WebSocket support

**Cons:**
- Not compatible with net/http
- Different programming model (Fasthttp-based)
- Smaller Go community adoption
- Some compatibility issues with standard Go libraries

**Reason not chosen:** Fiber's incompatibility with net/http ecosystem and different programming model would limit our ability to use standard Go libraries and tools.

### Standard Library (net/http)

**Pros:**
- No dependencies
- Maximum flexibility
- Guaranteed stability
- No framework lock-in

**Cons:**
- Requires significant boilerplate
- No built-in routing beyond basic patterns
- Manual implementation of common features
- Slower development speed

**Reason not chosen:** While viable for simple services, the lack of built-in features would significantly slow development and require reimplementing common patterns.

## Related

- ADR 002: RPC Framework
- ADR 003: Database Access Pattern
- ADR 004: Observability Stack
- [Gin Documentation](https://gin-gonic.com/docs/)
- [Go Web Framework Benchmark](https://github.com/smallnest/go-web-framework-benchmark)
- Design Doc: Microservice Template Architecture
