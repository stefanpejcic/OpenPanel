# Contributing to OpenPanel

Thank you for your interest in contributing to OpenPanel! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:
- Git
- Docker and Docker Compose
- Node.js (v16+)
- Python 3.12

### Setting Up the Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/OpenPanel.git
   cd OpenPanel
   ```

3. Set up the development environment:
   ```bash
   bash scripts/dev_setup.sh
   ```

## Development Workflow

### Branching Strategy

- `main` - Stable production-ready code
- `develop` - Development branch for the next release
- Feature branches should be created from `develop` and named according to the feature being developed

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Changes to the build process or auxiliary tools

Example: `feat: add support for MariaDB 10.6`

### Pull Request Process

1. Ensure your code passes all tests
2. Update documentation if needed
3. Submit a pull request to the `develop` branch
4. Wait for a maintainer to review your changes

## Testing

Run the test suite before submitting pull requests:

```bash
npm test      # For frontend components
pytest        # For Python backend components
bash scripts/integration_tests.sh  # For integration tests
```

## Documentation Guidelines

- Use plain language
- Include examples where possible
- Update relevant documentation when making code changes
- Follow Markdown best practices

## Installation Script Changes

When modifying the installation script (`install.sh`), please follow these guidelines:

1. **Variable Initialization**:
   - Initialize all variables at the top of the script
   - Use descriptive variable names in `UPPER_CASE` for global/environment variables
   - Document the purpose of each variable with a comment

2. **Function Organization**:
   - Group related functions together
   - Define functions before they are called
   - Document each function's purpose, parameters, and return values

3. **Error Handling**:
   - Use the `radovan()` function for consistent error reporting
   - Include detailed error messages to help troubleshooting
   - Check return codes after critical operations

4. **Logging**:
   - Use `debug_log()` for command execution that should be suppressed in normal mode
   - Use `log_message()` for structured logging with severity levels

5. **Testing**:
   - Test all changes on supported distributions (Ubuntu, Debian, AlmaLinux, etc.)
   - Verify both installation and uninstallation work properly
   - Test with different flags and configurations

6. **Code Style**:
   - Use consistent indentation (2 or 4 spaces)
   - Close all conditional blocks properly
   - Use meaningful variable and function names

Any changes to the installation script should be documented in both the code comments and the README.md Changelog section.

## Reporting Issues

When reporting issues, please include:

- A clear description of the problem
- Steps to reproduce
- Expected vs. actual behavior
- Screenshots if applicable
- Your environment details (OS, browser, etc.)

## Feature Requests

Feature requests are welcome! Please provide:

- A clear description of the feature
- Why it would be valuable to OpenPanel users
- Any design considerations or technical details

Thank you for contributing to OpenPanel!
