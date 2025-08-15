---
name: 📊 Observability Issue
about: Report issues with logging, metrics, tracing, or monitoring
title: '[OBS] '
labels: observability, triage
assignees: ''
---

## Issue Type
<!-- Select the type of observability issue -->

- [ ] Missing logs
- [ ] Incorrect metrics
- [ ] Broken traces
- [ ] Alert not firing
- [ ] Dashboard issue
- [ ] Performance degradation
- [ ] Other observability concern

## Description
<!-- Describe the observability issue -->

## Affected Components

### Service Details

- Service Name: 
- Version: 
- Environment: <!-- dev/staging/prod -->
- Time Range: <!-- When the issue occurred -->

### Observability Stack
- Logging: <!-- e.g., ELK, Loki -->
- Metrics: <!-- e.g., Prometheus -->
- Tracing: <!-- e.g., Jaeger, Tempo -->
- Dashboards: <!-- e.g., Grafana -->

## Expected vs Actual

### Expected Behavior
<!-- What should be logged/measured/traced -->

### Actual Behavior
<!-- What is actually happening -->

## Evidence
<!-- Provide evidence of the issue -->

### Log Samples
```
# Example logs showing the issue
```

### Metric Queries
```promql
# Prometheus queries demonstrating the issue
```

### Trace IDs
<!-- Provide example trace IDs if applicable -->
- 
- 

### Screenshots
<!-- Dashboard screenshots or other visual evidence -->

## Impact
<!-- How does this affect operations/debugging? -->

- [ ] 🔴 Critical - Cannot debug production issues
- [ ] 🟠 High - Significantly impacts operations
- [ ] 🟡 Medium - Causes inconvenience
- [ ] 🟢 Low - Minor issue

## Proposed Solution
<!-- Suggestions for fixing the issue -->

### Configuration Changes
```yaml
# Suggested config changes
```

### Code Changes
```go
// Suggested code changes
```

## Related Issues
<!-- Link to related issues or incidents -->

- Incident: 
- Related PR: 
- Documentation: 

## Checklist
- [ ] I have checked this isn't a duplicate issue
- [ ] I have provided specific examples
- [ ] I have included relevant time ranges
- [ ] I have suggested a solution if possible
