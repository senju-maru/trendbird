# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in TrendBird, please report it responsibly.

**Do NOT open a public GitHub Issue for security vulnerabilities.**

Instead, please send an email to the maintainers via GitHub's private vulnerability reporting:

1. Go to the [Security tab](../../security) of this repository
2. Click "Report a vulnerability"
3. Fill in the details

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix release**: Depends on severity, typically within 2 weeks for critical issues

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |

## Scope

The following are in scope:

- Backend API (Go / Connect-RPC)
- Frontend application (Next.js)
- Authentication flow (X OAuth 2.0 PKCE + JWT)
- Database queries (SQL injection)
- Environment variable handling

The following are out of scope:

- Vulnerabilities in third-party dependencies (please report to the upstream project)
- Social engineering attacks
- Denial of service attacks
