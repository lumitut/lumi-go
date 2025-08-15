---
name: 🐛 Bug Report
about: Report a bug or unexpected behavior
title: '[BUG] '
labels: bug, triage
assignees: ''
---

## Bug Description
<!-- A clear and concise description of what the bug is -->

## Environment

- **Service Version:** <!-- e.g., v1.2.3 or commit hash -->
- **Go Version:** <!-- e.g., 1.22 -->
- **Operating System:** <!-- e.g., Ubuntu 22.04, macOS 14.0 -->
- **Container Runtime:** <!-- e.g., Docker 24.0, containerd 1.7 -->
- **Kubernetes Version:** <!-- if applicable -->
- **Cloud Provider:** <!-- e.g., AWS, GCP, Azure, on-premise -->

## Steps to Reproduce
<!-- Steps to reproduce the behavior -->

1. Configure service with...
2. Send request to...
3. Observe...
4. See error

## Expected Behavior
<!-- A clear description of what you expected to happen -->

## Actual Behavior
<!-- What actually happened -->

## Error Messages/Logs
<!-- Include relevant error messages, stack traces, or logs -->

```
# Error output
```

<details>
<summary>Full logs</summary>

```
# Paste full logs here
```

</details>

## Request/Response Examples
<!-- If applicable, provide example requests and responses -->

### Request
```bash
curl -X POST http://localhost:8080/v1/endpoint \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

### Response
```json
{
  "error": "example error"
}
```

## Impact
<!-- Describe the impact of this bug -->

- [ ] 🔴 Critical - Service is down or data loss occurring
- [ ] 🟠 High - Major feature broken, no workaround
- [ ] 🟡 Medium - Feature broken, workaround available
- [ ] 🟢 Low - Minor issue, cosmetic

## Possible Solution
<!-- If you have suggestions on how to fix the bug -->

## Additional Context
<!-- Add any other context about the problem here -->

### Screenshots
<!-- If applicable, add screenshots to help explain your problem -->

### Related Issues
<!-- Link to related issues if any -->

## Checklist
- [ ] I have searched existing issues to ensure this isn't a duplicate
- [ ] I have included all relevant information
- [ ] I have provided steps to reproduce the issue
- [ ] I have included relevant logs and error messages
