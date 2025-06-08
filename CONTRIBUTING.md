# Contributing to pgtestkit

Thank you for considering contributing to Go TestDB! We welcome contributions from the community to help improve this project.

## How to Contribute

1. **Fork** the repository on GitHub
2. **Clone** the project to your own machine
3. **Commit** changes to your own branch
4. **Push** your work back up to your fork
5. Submit a **Pull Request** so that we can review your changes

## Development Setup

1. Make sure you have Go 1.21 or later installed
2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/pgtestkit.git
   cd pgtestkit
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests:
   ```bash
   go test -v ./...
   ```

## Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `gofmt` and `goimports` before committing
- Write tests for new features and bug fixes
- Update documentation when adding new features

## Reporting Issues

When reporting issues, please include:
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Go version and OS version
- Any relevant error messages or logs

## Feature Requests

We welcome feature requests! Please open an issue to discuss your idea before implementing it.

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
