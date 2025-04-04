# OpenPanel Security Policy

## Overview
Welcome to the OpenPanel Security Policy. This document outlines our approach to vulnerability management, reporting, and the types of security issues we consider significant.

Welcome and thanks for taking interest in OpenPanel!

We are mostly interested in reports by actual OpenPanel users but all high quality contributions are welcome.

If you believe that you have discovered a vulnerability in OpenPanel, please let our development team know by sending an email to <info@openpanel.com>
Please consider encrypting your report with our PGP key (available upon request) to ensure confidentiality.

## Responsible Disclosure
We appreciate responsible disclosure and ask that you provide a detailed description of the vulnerability, including:
- A list of affected services (e.g. nginx, opencli) and tested versions.
- Step-by-step reproduction instructions.
- Your findings and expected outcomes.

Please do not open any public issue on GitHub or any other social media until the report has been reviewed and a fix has been released.

With that, good luck hacking us ;)

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Qualifying Vulnerabilities

### Vulnerabilities we really care about

- Remote command execution
- Code/SQL Injection
- Authentication bypass
- Privilege Escalation
- Cross-site scripting (XSS)
- Performing limited admin actions without authorization
- CSRF
- Server-Side Request Forgery (SSRF)

### Vulnerabilities we accept

- Open redirects
- Password brute-forcing that circumvents rate limiting

## Non-Qualifying Vulnerabilities

- Theoretical attacks without proof of exploitability
- Attacks dependent on third-party libraries (report these to the maintainers)
- Social engineering
- Reflected file download
- Physical attacks
- Weak SSL/TLS/SSH algorithms or protocols
- Attacks involving physical access to a user’s device or compromised networks
- Attacks that consist solely of user input errors

## Documentation of Recent Updates
- The installation script now properly initializes the progress bar to avoid runtime issues.
- Fixed a syntax error in the installation script by adding a space before the file descriptor redirection (i.e. between ')' and '200>'), ensuring proper execution.
- Fixed several conditional statement closures and properly initialized variables to prevent "unbound variable" errors during installation.
- These enhancements support our ongoing commitment to secure and reliable operations.

## Additional Security Enhancements (v1.2.0)

### Authentication Improvements
- Implemented more robust password policies requiring minimum complexity standards
- Added support for two-factor authentication (2FA) for admin accounts
- Implemented automatic account lockout after multiple failed login attempts

### Web Application Security
- Upgraded CorazaWAF rules to protect against the latest OWASP Top 10 threats
- Added protection against XML External Entity (XXE) attacks
- Implemented Content Security Policy (CSP) headers by default
- Added protection against API abuse with rate limiting

### Infrastructure Security
- Improved container isolation for enhanced security between user environments
- Updated default firewall rules to block known malicious IP ranges
- Added automatic security scanning of Docker images before deployment
- Implemented secure TLS configuration with modern cipher suites only

## Security Best Practices for OpenPanel Administrators

### Regular Maintenance
1. **Keep OpenPanel Updated**: Always run the latest version to benefit from security patches.
2. **Monitor System Logs**: Set up regular log reviews to detect unusual activities.
3. **Perform Regular Backups**: Ensure automated backups are working and test restoration regularly.

### Hardening Recommendations
1. **Limit SSH Access**: Use SSH keys instead of passwords and consider changing the default SSH port.
2. **Implement Network Segmentation**: Isolate OpenPanel networks from other critical infrastructure.
3. **Enable Firewall**: Configure your firewall to allow only necessary services.
4. **Regular Security Audits**: Schedule periodic security reviews of your OpenPanel installation.

For detailed implementation instructions on these recommendations, please refer to our [Security Hardening Guide](https://openpanel.com/docs/admin/security/hardening).
