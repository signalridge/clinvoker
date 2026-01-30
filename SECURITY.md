# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisories**: Use [GitHub's private vulnerability reporting](https://github.com/signalridge/clinvoker/security/advisories/new)

2. **Email**: Send details to <security@signalridge.com>

### What to Include

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact assessment
- Any suggested fixes (if available)
- Your contact information for follow-up

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution Target**: Within 30 days (depending on complexity)

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt of your report
2. **Investigation**: We will investigate and validate the issue
3. **Updates**: We will keep you informed of our progress
4. **Resolution**: We will work on a fix and coordinate disclosure
5. **Credit**: We will credit you in the release notes (unless you prefer anonymity)

## Security Measures

### Code Security

- All code changes require review before merging
- Automated security scanning via `govulncheck`
- Dependency vulnerability scanning via Trivy
- Static analysis via golangci-lint

### Release Security

- All releases are built in CI with reproducible builds
- Checksums provided for all release artifacts
- SBOM (Software Bill of Materials) generated for releases

### Runtime Security

- No sensitive data stored in plain text
- Minimal filesystem access requirements
- Subprocess execution follows security best practices

## Scope

The following are in scope for security reports:

- Remote code execution
- Authentication/authorization bypass
- Information disclosure
- Privilege escalation
- Denial of service (with reasonable impact)
- Injection vulnerabilities

The following are out of scope:

- Issues requiring physical access to a device
- Social engineering attacks
- Issues in dependencies (report to upstream maintainers)
- Theoretical vulnerabilities without proof of concept

## Safe Harbor

We consider security research conducted in accordance with this policy to be:

- Authorized and lawful
- Helpful to the security of the project
- Protected from legal action on our part

We will not take legal action against researchers who:

- Act in good faith
- Do not access or modify user data
- Report findings responsibly
- Allow reasonable time for fixes before disclosure

Thank you for helping keep clinvk secure!
