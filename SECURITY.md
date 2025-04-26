# OpenPanel Security policy

Welcome and thanks for taking interest in OpenPanel!

We are mostly interested in reports by actual OpenPanel users but all high quality contributions are welcome.

If you believe that you have have discovered a vulnerability in OpenPanel, please let our development team know by sending an email to <info@openpanel.com>

We ask you to include a detailed description of the vulnerability, a list of services involved (e.g. nginx, opencli) and the versions which you've tested, full steps to reproduce the vulnerability, and include your findings and expected results.

Please do not open any public issue on Github or any other social media before the report has been published and a fix has been released.

With that, good luck hacking us ;)

## Supported versions

| Version | Supported          |
| ------- | ------------------ |
| [Latest](https://openpanel.com/docs/changelog/intro/#latest)  | :white_check_mark: |

## Qualifying Vulnerabilities

### Vulnerabilities we really care about

- Remote command execution
- Code/SQL Injection
- Authentication bypass
- Privilege Escalation
- Cross-site scripting (XSS)
- Performing limited admin actions without authorization
- CSRF

### Vulnerabilities we accept

- Open redirects
- Password brute-forcing that circumvents rate limiting

## Non-Qualifying Vulnerabilities

- Theoretical attacks without proof of exploitability
- Attacks that are the result of a third party library should be reported to the library maintainers
- Social engineering
- Reflected file download
- Physical attacks
- Weak SSL/TLS/SSH algorithms or protocols
- Attacks involving physical access to a user’s device, or involving a device or network that’s already seriously compromised (eg man-in-the-middle).
- The user attacks themselves
