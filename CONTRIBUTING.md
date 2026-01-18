# Contributing to terraform-provider-danubedata

Thank you for your interest in contributing to the DanubeData Terraform Provider!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Make
- A DanubeData account (for acceptance tests)

### Building

```bash
# Clone the repository
git clone https://github.com/AdrianSilaghi/terraform-provider-danubedata.git
cd terraform-provider-danubedata

# Install dependencies
make deps

# Build the provider
make build

# Install locally for testing
make install
```

### Running Tests

```bash
# Run unit tests
make test

# Run unit tests with coverage
make test-coverage

# Run acceptance tests (requires API token)
export DANUBEDATA_API_TOKEN="your-test-token"
make testacc

# Run specific acceptance tests
make testacc-Vps
make testacc-SshKey
```

## Making Changes

### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Run `make lint` to check for issues
- Add tests for new functionality

### Commit Messages

Use conventional commit format:

```
feat: add new resource danubedata_widget
fix: correct timeout handling in VPS creation
docs: update README with new examples
test: add acceptance tests for cache resource
chore: update dependencies
```

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-new-feature`
3. Make your changes
4. Run checks: `make check`
5. Commit your changes
6. Push to your fork
7. Open a pull request

### Adding a New Resource

1. Create the client methods in `internal/client/`
2. Create the resource in `internal/resources/`
3. Register the resource in `internal/provider/provider.go`
4. Add unit tests for client methods
5. Add acceptance tests for the resource
6. Add documentation in `docs/resources/`
7. Add an example in `examples/`

### Adding a New Data Source

1. Create the client methods (if needed) in `internal/client/`
2. Create the data source in `internal/datasources/`
3. Register the data source in `internal/provider/provider.go`
4. Add tests
5. Add documentation in `docs/data-sources/`

## Testing Guidelines

### Unit Tests

Unit tests use mock HTTP servers to test client functionality:

```go
func TestClient_CreateWidget(t *testing.T) {
    server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        // Return mock response
    })
    defer server.Close()

    c := newTestClient(server)
    // Test client method
}
```

### Acceptance Tests

Acceptance tests create real resources. They:

- Require `TF_ACC=1` environment variable
- Require `DANUBEDATA_API_TOKEN` environment variable
- Use random names to avoid conflicts
- Clean up resources after tests

```go
func TestAccWidgetResource_basic(t *testing.T) {
    name := acctest.RandomName("tf-widget")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acctest.PreCheck(t) },
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccWidgetConfig(name),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("danubedata_widget.test", "name", name),
                ),
            },
        },
    })
}
```

## Documentation

- Resource documentation goes in `docs/resources/<name>.md`
- Data source documentation goes in `docs/data-sources/<name>.md`
- Examples go in `examples/<use-case>/main.tf`
- Update CHANGELOG.md for notable changes

## Questions?

Open an issue if you have questions or need help!
