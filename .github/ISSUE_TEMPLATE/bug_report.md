---
name: Bug Report
about: Report a bug in the lumi-go microservice
title: '[BUG] '
labels: 'bug, needs-triage'
assignees: ''
---

## Bug Description
<!-- Provide a clear and concise description of the bug -->

## Environment
<!-- Please complete the following information -->

- **OS**: [e.g., Ubuntu 22.04, macOS 14.0, Windows 11]
- **Go Version**: [e.g., 1.22.0]
- **Service Version/Commit**: [e.g., v1.0.0 or commit hash]
- **Deployment Method**: [e.g., Docker, Kubernetes, Binary]

## Steps to Reproduce
<!-- Provide detailed steps to reproduce the behavior -->

1. Start the service with configuration: `...`
2. Send request to endpoint: `...`
3. Observe error: `...`

## Expected Behavior
<!-- Describe what you expected to happen -->

## Actual Behavior
<!-- Describe what actually happened -->

## Configuration
<!-- Include relevant configuration (remove sensitive data) -->

```json
{
  "service": {
    "name": "lumi-go",
    "environment": "development"
  },
  "clients": {
    "database": {
      "enabled": false
    }
  }
  // ... relevant config
}
```

## Logs
<!-- Include relevant log output -->

```
[timestamp] ERROR [correlation_id] Error message here
...
```

## Stack Trace
<!-- If applicable, include the full stack trace -->

```
panic: runtime error...
goroutine 1 [running]:
...
```

## Request/Response Examples
<!-- If it's an API issue, provide request and response examples -->

### Request
```bash
curl -X POST http://localhost:8080/api/endpoint \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'
```

### Response
```json
{
  "error": "error message",
  "code": 500
}
```

## Additional Context
<!-- Add any other context about the problem here -->

## Possible Solution
<!-- If you have ideas on how to fix this, please share -->

## Workaround
<!-- If you found a workaround, please describe it to help others -->

---
**For Maintainers:**
- [ ] Bug confirmed
- [ ] Root cause identified
- [ ] Test case added
- [ ] Fix implemented
- [ ] Documentation updated if needed
