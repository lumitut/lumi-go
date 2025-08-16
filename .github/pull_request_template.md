## Description
<!-- Provide a brief description of the changes in this PR -->

## Type of Change
<!-- Mark the relevant option with an "x" -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Code style/refactoring
- [ ] ⚡ Performance improvement
- [ ] 🔧 Configuration change
- [ ] 🧪 Test improvement
- [ ] 🔒 Security fix

## Related Issue
<!-- Link to the related issue (e.g., Fixes #123, Closes #456) -->

Fixes #

## Changes Made
<!-- List the specific changes made in this PR -->

- 
- 
- 

## Testing
<!-- Describe the tests you ran to verify your changes -->

- [ ] Unit tests pass (`make test-unit`)
- [ ] Integration tests pass (`make test-integration`)
- [ ] Linting passes (`make lint`)
- [ ] Local testing completed

### Test Coverage
<!-- Include test coverage if applicable -->
```
Current coverage: XX%
```

## Checklist
<!-- Mark completed items with an "x" -->

### Code Quality
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings or errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

### Documentation
- [ ] I have updated the README if needed
- [ ] I have updated API documentation if applicable
- [ ] I have added/updated code comments where necessary
- [ ] I have updated CHANGELOG.md if applicable

### Configuration
- [ ] I have updated `cmd/server/schema/lumi.json` if adding new config options
- [ ] I have updated `env.example` if adding new environment variables
- [ ] Configuration changes are backward compatible

### Dependencies
- [ ] I have run `go mod tidy` if dependencies changed
- [ ] No unnecessary dependencies were added
- [ ] Security scan passes for new dependencies

### Performance
- [ ] My changes don't negatively impact performance
- [ ] I have run benchmarks if applicable

## Screenshots/Logs
<!-- If applicable, add screenshots or logs to help explain your changes -->

## Additional Notes
<!-- Add any additional notes or context about the PR here -->

## Reviewer Guidelines
<!-- Help reviewers understand what to focus on -->

Please pay special attention to:
- [ ] Error handling
- [ ] Resource cleanup (defer statements)
- [ ] Concurrent access safety
- [ ] Configuration validation
- [ ] Test coverage

---
**PR Readiness Checklist for Reviewers:**
- [ ] Code is clean and follows Go best practices
- [ ] Tests are comprehensive and pass
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact is acceptable
