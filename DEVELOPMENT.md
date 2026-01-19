# Development Guide

This document provides detailed information for developers working on the Liqo Security Engine.

## Prerequisites

### Required Tools

- **Go**: 1.24.6 or higher
- **Docker**: 17.03 or higher
- **kubectl**: 1.11.3 or higher
- **kind** (recommended) or access to a Kubernetes cluster
- **make**: For running build tasks
- **golangci-lint**: For code linting

### Optional Tools

- **kubebuilder**: For making changes to the API
- **controller-gen**: For generating CRD manifests (included with kubebuilder)
- **kustomize**: For customizing Kubernetes manifests
- **Helm**: For working with the Helm chart

## Setting Up Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/riccardotornesello/liqo-security-engine.git
cd liqo-security-engine
```

### 2. Install Liqo (Development Version)

This project has a `replace` directive in `go.mod` that points to a local Liqo installation:

```go
replace github.com/liqotech/liqo => ../liqo
```

**Development Setup**:

```bash
# Clone Liqo alongside this project
cd ..
git clone https://github.com/liqotech/liqo.git
cd liqo-security-engine
```

**For Production/Release**:

Before releasing, update `go.mod` to use a specific Liqo version:

```go
require (
    github.com/liqotech/liqo v1.0.3
    // ... other dependencies
)

// Remove or comment out the replace directive
// replace github.com/liqotech/liqo => ../liqo
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Install CRDs

```bash
make install
```

### 5. Set Up a Test Cluster

#### Using kind

```bash
# Create a kind cluster
kind create cluster --name liqo-dev

# Install Liqo
# Follow the Liqo installation guide: https://docs.liqo.io/installation/

# Ensure Liqo is running
kubectl get pods -n liqo
```

## Development Workflow

### Building the Project

```bash
# Build the binary
make build

# The binary will be in ./bin/manager
./bin/manager --help
```

### Running Locally

To run the controller locally against your Kubernetes cluster:

```bash
# Install CRDs
make install

# Run the controller (uses your current kubectl context)
make run
```

The controller will start and connect to your cluster. Press Ctrl+C to stop it.

### Making Code Changes

#### 1. Update the Code

Make your changes to the Go source files.

#### 2. Run Tests

```bash
# Run unit tests
make test

# Run specific tests
go test ./internal/controller/... -v

# Run with coverage
make test-coverage
```

#### 3. Lint Your Code

```bash
# Run golangci-lint
make lint

# Fix auto-fixable issues
make lint-fix
```

#### 4. Format Code

```bash
# Format all Go files
go fmt ./...

# Or use gofmt with write flag
gofmt -w .
```

### Changing the API

If you modify the API types in `api/v1/`:

```bash
# Generate DeepCopy methods
make generate

# Generate CRD manifests
make manifests

# Test that everything compiles
make build

# Update examples if needed
# Edit files in examples/

# Re-install CRDs
make install
```

### Building Container Images

```bash
# Build the Docker image
make docker-build IMG=<your-registry>/liqo-security-engine:dev

# Push the image
make docker-push IMG=<your-registry>/liqo-security-engine:dev
```

### Deploying to a Cluster

```bash
# Deploy using your custom image
make deploy IMG=<your-registry>/liqo-security-engine:dev

# Check deployment
kubectl get deployment -n liqo-system
kubectl logs -n liqo-system -l app=liqo-security-engine
```

### Testing Your Changes

#### Unit Tests

Add tests in `*_test.go` files next to your code:

```go
func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := MyFunction(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

#### Integration Tests

Integration tests are in the `test/` directory:

```bash
# Run integration tests
make test-integration
```

#### E2E Tests

End-to-end tests require a running cluster with Liqo:

```bash
# Run e2e tests
make test-e2e
```

#### Manual Testing

1. Deploy a test PeeringConnectivity resource:
   ```bash
   kubectl apply -f examples/provider.yaml
   ```

2. Check the status:
   ```bash
   kubectl get peeringconnectivity -A
   kubectl describe peeringconnectivity <name> -n <namespace>
   ```

3. Verify FirewallConfiguration was created:
   ```bash
   kubectl get firewallconfiguration -A
   ```

4. Test network connectivity between pods

## Debugging

### Debug Mode

Run with verbose logging:

```bash
make run ARGS="--zap-log-level=debug"
```

### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/main.go -- --zap-log-level=debug
```

### Debugging in VS Code

Add this to `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Controller",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/main.go",
            "args": ["--zap-log-level=debug"]
        }
    ]
}
```

### Viewing Controller Logs

```bash
# If running locally
# Logs appear in the terminal

# If deployed to cluster
kubectl logs -n liqo-system -l app=liqo-security-engine -f
```

## Project Structure

```
.
├── api/v1/                          # API definitions (CRDs)
│   ├── groupversion_info.go         # API group registration
│   ├── peeringsecurity_types.go     # PeeringConnectivity type
│   └── zz_generated.deepcopy.go     # Generated deepcopy methods
├── cmd/                             # Main application
│   └── main.go                      # Entry point
├── config/                          # Kubernetes manifests
│   ├── crd/                         # CRD definitions
│   ├── manager/                     # Controller deployment
│   ├── rbac/                        # RBAC configurations
│   ├── samples/                     # Sample resources
│   └── ...                          # Other configs
├── dist/                            # Distribution artifacts
│   ├── chart/                       # Helm chart
│   └── install.yaml                 # Combined manifest
├── examples/                        # Usage examples
├── hack/                            # Build scripts
├── internal/                        # Internal packages
│   └── controller/                  # Controller implementation
│       ├── forge/                   # Firewall config generation
│       ├── utils/                   # Utility functions
│       └── peeringsecurity_controller.go
├── test/                            # Tests
│   ├── e2e/                         # End-to-end tests
│   └── utils/                       # Test utilities
├── Dockerfile                       # Container image definition
├── Makefile                         # Build automation
├── go.mod                           # Go module definition
└── ...                              # Documentation files
```

## Common Make Targets

```bash
make help              # Show all available targets
make build             # Build the binary
make run               # Run locally
make test              # Run tests
make lint              # Run linter
make generate          # Generate code
make manifests         # Generate manifests
make install           # Install CRDs
make uninstall         # Uninstall CRDs
make deploy            # Deploy to cluster
make undeploy          # Remove from cluster
make docker-build      # Build container image
make docker-push       # Push container image
make build-installer   # Build install.yaml
```

## Code Style Guidelines

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use meaningful names
- Keep functions focused and small
- Write tests for new code
- Document exported types and functions

### Comments

- Package comments on every package
- Function comments on exported functions
- Explain "why" not just "what"
- Use complete sentences

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("unable to get pod: %w", err)
}

// Bad
if err != nil {
    return err
}
```

## Troubleshooting

### CRD Already Exists

```bash
make uninstall
make install
```

### Controller Not Starting

1. Check RBAC permissions
2. Verify CRDs are installed
3. Check controller logs
4. Ensure cluster is accessible

### Tests Failing

1. Ensure test cluster is running
2. Check if CRDs are installed
3. Verify Liqo is properly installed
4. Check test logs for specific errors

### Docker Build Fails

1. Ensure Docker daemon is running
2. Check Dockerfile for errors
3. Verify base image is accessible
4. Check for sufficient disk space

## Getting Help

- Check [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
- Review [TODOS.md](TODOS.md) for known issues
- Search [existing issues](https://github.com/riccardotornesello/liqo-security-engine/issues)
- Ask in [discussions](https://github.com/riccardotornesello/liqo-security-engine/discussions)

## Release Process

See [RELEASING.md](RELEASING.md) for information about the release process.
