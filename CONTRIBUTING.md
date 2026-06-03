# Contributing to wfh

Thank you for considering contributing to wfh. This project prioritizes
privacy, simplicity, and reliability above all else.

## Privacy Commitment

wfh is **privacy-first** by design. Contributions should never introduce:

- Telemetry or analytics of any kind
- Cloud sync or remote data transmission
- Third-party API calls
- User tracking or fingerprinting
- External dependencies on cloud services

## Development

```bash
# Clone and build
git clone https://github.com/zinuo-xu/wfh.git
cd wfh
make build

# Run tests
make test

# Run linter
make lint
```

## Pull Request Process

1. Ensure no privacy-compromising code is introduced
2. Update tests if adding functionality
3. Update documentation for user-facing changes
4. Run `make lint` and `make test` before submitting
5. Keep changes focused — one feature per PR

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Document exported functions with Go-style comments
- Keep internal packages free of external dependencies where possible
- Add privacy-first notes to any new tracking features

## Reporting Issues

Include your OS, Go version, and steps to reproduce. Never include
activity data or screenshots containing personal information.

## License

By contributing, you agree that your contributions will be licensed
under the MIT License.
