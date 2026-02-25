---
name: cli-local-build
description: Build and test the Kurtosis CLI from source. Compile the CLI binary locally, run it against Docker or Kubernetes engines, and iterate on CLI changes without creating a release. Use when developing or debugging CLI commands.
compatibility: Requires Go 1.23+.
metadata:
  author: ethpandaops
  version: "1.0"
---

# CLI Local Build

Build and test the Kurtosis CLI from source for local development.

## Quick build

```bash
# Build the CLI binary
go build -o /tmp/kurtosis ./cli/cli/

# Verify it works
/tmp/kurtosis version
```

## Using the local CLI

The locally built CLI works exactly like the installed one. Use the full path to avoid conflicts with the system-installed `kurtosis`:

```bash
# Start engine
/tmp/kurtosis engine start

# Run a package
/tmp/kurtosis run github.com/ethpandaops/ethereum-package

# Clean up
/tmp/kurtosis clean -a

# Stop engine
/tmp/kurtosis engine stop
```

## Switching between Docker and Kubernetes

```bash
# Check current cluster setting
/tmp/kurtosis cluster get

# Switch to Docker
/tmp/kurtosis cluster set docker

# Switch to Kubernetes (uses current kubectl context)
/tmp/kurtosis cluster set kubernetes

# Restart engine after switching
/tmp/kurtosis engine restart
```

## Build with race detector

For debugging concurrency issues:

```bash
go build -race -o /tmp/kurtosis ./cli/cli/
```

## Running tests

```bash
# Run CLI command tests
go test ./cli/cli/commands/...

# Run a specific test
go test -run TestName ./cli/cli/commands/...

# Run with verbose output
go test -v ./cli/cli/commands/...
```

## Key source locations

| Component | Path |
|-----------|------|
| CLI entry point | `cli/cli/main.go` |
| CLI commands | `cli/cli/commands/` |
| Engine launcher | `engine/launcher/` |
| API container launcher | `core/launcher/` |
| Container engine abstraction | `container-engine-lib/` |
| gRPC API definitions | `api/protobuf/` |
| Version constant | `kurtosis_version/kurtosis_version.go` |

## Module dependency order

The monorepo has multiple Go modules. If you change a dependency, rebuild in order:

```
container-engine-lib
  → contexts-config-store
    → grpc-file-transfer
      → name-generator
        → api
          → metrics-library
            → engine
              → core
                → cli
```

Most CLI-only changes just need `go build ./cli/cli/`.

## Common issues

- **`go build` fails with import errors**: Run `go mod tidy` in the failing module directory
- **CLI shows wrong version**: The version comes from `kurtosis_version/kurtosis_version.go` — it's compiled into the binary
- **Engine image mismatch**: The CLI pulls engine images matching its compiled version. For dev testing with custom images, see the `k8s-dev-deploy` skill
