# Secret Scanning Configuration

## Overview

This document outlines how to enable and configure secret scanning for the lumi-go (Go Middle-Service Template) repository.

## GitHub Secret Scanning Features

### 1. Enable Secret Scanning (Repository Admin Required)

Navigate to your repository settings and enable the following features:

1. **Settings → Code security and analysis**
2. Enable **Secret scanning** 
   - Detects tokens, private keys, and other secrets in your repository
   - Scans all branches and the entire Git history
3. Enable **Push protection** 
   - Prevents commits containing secrets from being pushed
   - Blocks pushes at the Git level before secrets enter the repository

### 2. Configure Secret Scanning Alerts

1. **Settings → Code security and analysis → Secret scanning**
2. Configure alert notifications:
   - Email notifications to security team
   - Create issues for detected secrets
   - Webhook integration with security tools

### 3. Custom Pattern Configuration

Create custom patterns for organization-specific secrets:

```yaml
# Example custom patterns (add via Settings → Code security → Secret scanning)
patterns:
  - name: "Lumitut API Key"
    pattern: "lumi_[a-zA-Z0-9]{32}"
    
  - name: "Internal JWT"
    pattern: "eyJ[a-zA-Z0-9_-]+\\.eyJ[a-zA-Z0-9_-]+\\.[a-zA-Z0-9_-]+"
```

## Local Pre-commit Scanning

### Install gitleaks for local secret detection:

```bash
# macOS
brew install gitleaks

# Linux
wget https://github.com/gitleaks/gitleaks/releases/download/v8.18.1/gitleaks_8.18.1_linux_x64.tar.gz
tar -xzf gitleaks_8.18.1_linux_x64.tar.gz

# Or via Go
go install github.com/gitleaks/gitleaks/v8/cmd/gitleaks@latest
```

### Run manual scan:

```bash
# Scan repository
gitleaks detect --source . -v

# Scan specific commit
gitleaks detect --source . --log-opts="HEAD^..HEAD"

# Scan with custom config
gitleaks detect --source . --config .gitleaks.toml
```

## Supported Secret Types

GitHub automatically scans for:

- Cloud provider credentials (AWS, Azure, GCP)
- API keys from popular services
- Database connection strings
- Private keys (SSH, GPG, SSL certificates)
- OAuth tokens and refresh tokens
- Webhook URLs with embedded tokens
- Package registry tokens

## Response Procedures

### When a Secret is Detected:

1. **Immediate Actions:**
   - Revoke the exposed credential immediately
   - Generate a new credential
   - Update all systems using the credential

2. **Repository Cleanup:**
   - Remove the secret from the current branch
   - If in history, consider using `git filter-branch` or BFG Repo-Cleaner
   - Force push if necessary (coordinate with team)

3. **Documentation:**
   - Document the incident
   - Update security procedures if needed
   - Review how the secret was exposed

### Prevention Best Practices:

1. **Never commit secrets directly**
   - Use environment variables
   - Use secret management services (AWS Secrets Manager, Vault)
   - Use `.env` files (git-ignored) for local development

2. **Use placeholders in code:**
   ```go
   // Good
   apiKey := os.Getenv("API_KEY")
   
   // Bad
   apiKey := "sk_live_abc123xyz789"
   ```

3. **Review before committing:**
   - Use `git diff --staged` before committing
   - Enable pre-commit hooks
   - Review PR changes carefully

## Secret Management Architecture

For production systems, follow this hierarchy:

1. **Production:** AWS Secrets Manager / HashiCorp Vault
2. **Staging:** AWS Secrets Manager with different IAM roles
3. **Development:** Environment variables from `.env` files
4. **CI/CD:** GitHub Secrets / AWS IAM roles

## Monitoring and Compliance

- Review secret scanning alerts weekly
- Audit secret access logs monthly
- Rotate all secrets quarterly
- Maintain secret inventory documentation

## Additional Tools

### Recommended Security Tools:

- **truffleHog:** Deep Git history scanning
- **detect-secrets:** Pre-commit framework integration
- **git-secrets:** AWS-focused secret scanning

### Integration with CI/CD:

The `security-scan.yml` workflow includes secret scanning as part of the CI pipeline. Results are uploaded to GitHub Security tab for centralized viewing.

## Contact

For security concerns or questions:
- Security Team: security@lumitut.com
- Platform Team: platform@lumitut.com

## References

- [GitHub Secret Scanning Documentation](https://docs.github.com/en/code-security/secret-scanning)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [CWE-798: Use of Hard-coded Credentials](https://cwe.mitre.org/data/definitions/798.html)
