# Enable GitHub Security Features

## Quick Setup Guide

After creating a repository from this template, repository administrators must manually enable the following GitHub security features:

### 1. Navigate to Repository Settings

Go to: **Settings** → **Code security and analysis**

### 2. Enable Security Features

Enable the following features by clicking their respective "Enable" buttons:

- [x] **Dependency graph** - Understand your dependencies
- [x] **Dependabot alerts** - Get notified about vulnerabilities
- [x] **Dependabot security updates** - Automatic security PRs
- [x] **Secret scanning** - Detect secrets in code
- [x] **Push protection** - Block commits with secrets
- [x] **CodeQL analysis** - Advanced code scanning (already configured via workflow)

### 3. Configure Additional Settings

#### Secret Scanning Alerts
- Click on **Secret scanning** settings
- Configure notification preferences
- Add custom patterns if needed (see SECRET_SCANNING.md)

#### Branch Protection Rules
- Go to **Settings** → **Branches**
- Add rule for `main` branch:
  - [x] Require pull request reviews (1 minimum)
  - [x] Dismiss stale reviews
  - [x] Require status checks (CI, security-scan)
  - [x] Require branches to be up to date
  - [x] Include administrators
  - [x] Require conversation resolution

#### Security Advisories
- Go to **Security** → **Advisories**
- Enable private vulnerability reporting

### 4. Team Configuration

Update the following references in the configuration files:

1. In `.github/dependabot.yml`:
   - Replace `@lumitut/platform-team` with your platform team handle
   - Replace `@lumitut/security-team` with your security team handle

2. In `SECURITY.md`:
   - Update email addresses for security contacts
   - Update emergency hotline if applicable

### 5. Repository Secrets

Add the following secrets for CI/CD (Settings → Secrets and variables → Actions):

- `AWS_ACCOUNT_ID` - AWS account for ECR
- `AWS_REGION` - AWS region (e.g., us-east-1)
- `SONAR_TOKEN` - If using SonarCloud (optional)
- `SLACK_WEBHOOK` - For security notifications (optional)

### 6. Verify Setup

Run the following checks:

```bash
# Check Dependabot is working
# You should see Dependabot PRs within 24 hours if there are any updates

# Test secret scanning locally
gitleaks detect --source . -v

# Run security workflow manually
# Go to Actions → Security Scanning → Run workflow

# Verify branch protection
# Try pushing directly to main (should be blocked)
```

### 7. Post-Setup Tasks

- [ ] Schedule security review meeting
- [ ] Document custom secret patterns
- [ ] Set up security alert notifications
- [ ] Train team on security workflow
- [ ] Create security runbook

## Automation Note

While most security features require manual enablement through the GitHub UI, the following are already configured via files in this template:

- ✅ Dependabot configuration (`.github/dependabot.yml`)
- ✅ Security scanning workflows (`.github/workflows/security-scan.yml`)
- ✅ Secret scanning configuration (`.gitleaks.toml`)
- ✅ Pre-commit hooks (`.pre-commit-config.yaml`)
- ✅ Security policy (`SECURITY.md`)

## Support

For questions or issues with security setup:
- Create an issue in this repository
- Contact the platform team
- Refer to the [GitHub Security Documentation](https://docs.github.com/en/code-security)
