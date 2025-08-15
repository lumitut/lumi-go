# Security Policy

## Overview

This document outlines the security policy for the Go Middle-Service Template (GMT) and services built using this template. Security is a critical aspect of our microservices architecture, and we follow defense-in-depth principles.

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          | Notes                                    |
| ------- | ------------------ | ---------------------------------------- |
| 1.x.x   | :white_check_mark: | Current stable release                  |
| 0.x.x   | :warning:          | Pre-release, security fixes case-by-case |
| < 0.0.11| :x:               | No longer supported                      |

## Security Features

### Built-in Security Controls

1. **Authentication & Authorization**
   - JWT-based authentication with JWKS support
   - Role-based access control (RBAC)
   - Optional mTLS for service-to-service communication
   - Token expiration and refresh mechanisms

2. **Input Validation**
   - Request payload validation using `go-playground/validator`
   - SQL injection prevention via prepared statements
   - XSS protection through proper escaping
   - Rate limiting per client/route

3. **Transport Security**
   - TLS 1.2+ enforcement
   - HTTPS-only in production
   - Secure headers middleware
   - CORS configuration with strict defaults

4. **Data Protection**
   - Encryption at rest for sensitive data
   - Secure secret management (no hardcoded secrets)
   - PII redaction in logs
   - Audit logging for sensitive operations

5. **Dependency Security**
   - Automated dependency scanning via Dependabot
   - Container image vulnerability scanning
   - License compliance checking
   - SBOM generation for supply chain transparency

## Security Scanning

This template includes multiple layers of security scanning:

### Automated Scanning

- **Dependabot**: Daily checks for vulnerable dependencies
- **CodeQL**: Static analysis for security vulnerabilities
- **Gosec**: Go-specific security analysis
- **Trivy**: Container image vulnerability scanning
- **Gitleaks**: Secret detection in code and git history

### Manual Security Review

- Code review required for all PRs
- Security team review for authentication/authorization changes
- Penetration testing for production deployments
- Regular security audits (quarterly)

## Vulnerability Reporting

### Reporting Process

1. **DO NOT** create public GitHub issues for security vulnerabilities
2. Email security details to: **security@lumitut.com**
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 5 business days
- **Resolution Target**: 
  - Critical: 7 days
  - High: 14 days
  - Medium: 30 days
  - Low: 90 days

### Responsible Disclosure

We support responsible disclosure and will:
- Keep you informed about the progress
- Credit you for the discovery (unless you prefer anonymity)
- Not pursue legal action for good-faith security research

## Security Checklist for Developers

### Before Committing Code

- [ ] No hardcoded secrets or credentials
- [ ] All user inputs are validated
- [ ] SQL queries use prepared statements
- [ ] Error messages don't leak sensitive information
- [ ] Logging doesn't include PII or secrets
- [ ] Dependencies are up to date

### Before Deploying

- [ ] Security scans pass (no HIGH/CRITICAL vulnerabilities)
- [ ] Environment variables are properly configured
- [ ] TLS certificates are valid and not expiring soon
- [ ] Rate limiting is configured appropriately
- [ ] Monitoring and alerting are set up
- [ ] Audit logging is enabled

### Production Security

- [ ] Principle of least privilege for service accounts
- [ ] Network policies restrict unnecessary communication
- [ ] Secrets are managed via secret management system
- [ ] Regular security updates are applied
- [ ] Incident response plan is documented and tested
- [ ] Backup and disaster recovery procedures are in place

## Security Tools and Commands

### Run Security Scans Locally

```bash
# Go vulnerability check
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Static security analysis
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Secret scanning
gitleaks detect --source . -v

# Container scanning (after building)
trivy image lumitut/gmt-template:latest

# License compliance
go-licenses check ./...
```

### Security-Related Make Targets

```bash
make security-scan    # Run all security scans
make vuln-check      # Check for known vulnerabilities
make secret-scan     # Scan for secrets
make container-scan  # Scan container image
make audit          # Full security audit
```

## Compliance and Standards

This template is designed to help meet common compliance requirements:

- **OWASP Top 10**: Protections against common web vulnerabilities
- **CIS Benchmarks**: Container and Kubernetes security best practices
- **PCI DSS**: Secure coding practices for payment processing
- **GDPR**: Data protection and privacy controls
- **SOC 2**: Security controls for service organizations

## Security Contacts

- **Security Team**: security@lumitut.com
- **Platform Team**: platform@lumitut.com
- **Emergency Hotline**: +1-XXX-XXX-XXXX (Critical issues only)

## Security Training Resources

- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet.html)
- [Go Security Best Practices](https://github.com/guardrailsio/awesome-golang-security)
- [Container Security Best Practices](https://sysdig.com/learn-cloud-native/kubernetes-security/container-security-best-practices/)
- Internal Security Training: Available on the company wiki

## Change Log

| Date       | Version | Changes                                      |
|------------|---------|----------------------------------------------|
| 2024-08-15 | 1.0.0   | Initial security policy                     |
| TBD        | 1.1.0   | Added container scanning and SBOM generation |

## Acknowledgments

We thank the security researchers and community members who have helped improve our security posture through responsible disclosure.

---

*This security policy is a living document and will be updated as our security practices evolve.*
