# Contributing to WhatsApp CLI

First off, thanks for taking the time to contribute! 🎉

We welcome contributions from everyone. By participating in this project, you help make it better for everyone.

## How to Contribute

### Reporting Bugs

If you find a bug, please create a new issue and include:
- A clear description of the issue.
- Steps to reproduce the bug.
- Your operating system and architecture (e.g., Windows 10, macOS M1, Ubuntu 22.04).
- Logs or screenshots if applicable.

### Suggesting Enhancements

Have an idea for a new feature?
- Check if the feature has already been suggested in the Issues.
- If not, create a new issue using the "Feature Request" template (if available) or just describe your idea clearly.

### Coding Guidelines

- **Go Version**: Ensure you are using Go 1.24 or later.
- **Formatting**: Run `gofmt` or `go fmt ./...` before committing. Resulting code should differ only in ways that are stylistic and not semantic.
- **Code Style**: Try to follow standard Go idioms (Effective Go).
- **Dependencies**: Use `go mod tidy` to manage dependencies.

### Pull Request Process

1.  **Fork** the repository to your own GitHub account.
2.  **Clone** individual fork to your local machine:
    ```bash
    git clone https://github.com/YOUR_USERNAME/WhatsApp-CLI.git
    ```
3.  **Create a branch** for your feature or fix:
    ```bash
    git checkout -b feature/amazing-feature
    ```
4.  **Make your changes** and commit them with a clear message:
    ```bash
    git commit -m "feat: Add amazing feature"
    ```
5.  **Push** to your fork:
    ```bash
    git push origin feature/amazing-feature
    ```
6.  Open a **Pull Request** on the main repository.

## Development Setup

1.  Install [Go](https://go.dev/dl/) (1.24+).
2.  Clone the project.
3.  Install dependencies:
    ```bash
    go mod tidy
    ```
4.  Run the project locally:
    ```bash
    go run main.go
    ```

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
