---
name: run-package
description: Run Starlark scripts and packages with kurtosis run. Covers all flags including dry-run, args-file, parallel execution, image download modes, verbosity levels, and production mode. Use when executing Kurtosis packages locally or from GitHub.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Run Package

Execute Starlark scripts and packages with `kurtosis run`.

## Basic usage

```bash
# Run a local package
kurtosis run ./my-package

# Run a local .star script
kurtosis run ./script.star

# Run a remote package from GitHub
kurtosis run github.com/ethpandaops/ethereum-package

# Run with inline args
kurtosis run github.com/ethpandaops/ethereum-package '{"participants": [{"el_type": "geth", "cl_type": "lighthouse"}]}'

# Run with args from a file (JSON or YAML)
kurtosis run github.com/ethpandaops/ethereum-package --args-file config.yaml
```

## Named enclaves

```bash
# Run in a specific enclave (created if it doesn't exist)
kurtosis run --enclave my-testnet github.com/ethpandaops/ethereum-package

# Re-run in an existing enclave (adds to it)
kurtosis run --enclave my-testnet ./additional-services.star
```

## Dry run

Preview what will execute without making changes:

```bash
kurtosis run --dry-run github.com/ethpandaops/ethereum-package --args-file config.yaml
```

## Verbosity levels

```bash
# Default — concise description of what happens
kurtosis run ./my-package

# Brief — concise but explicit
kurtosis run -v brief ./my-package

# Detailed — all arguments for each instruction
kurtosis run -v detailed ./my-package

# Executable — generates copy-pasteable Starlark
kurtosis run -v executable ./my-package

# Output only — just the return value
kurtosis run -v output_only ./my-package
```

## Image handling

```bash
# Default: only pull if image doesn't exist locally
kurtosis run ./my-package

# Always pull latest image tags
kurtosis run --image-download always ./my-package
```

## Parallel execution

```bash
# Run instructions in parallel (as soon as dependencies resolve)
kurtosis run --parallel ./my-package

# Set parallelism level
kurtosis run --parallel --parallelism 8 ./my-package
```

## Advanced options

```bash
# Production mode — services auto-restart on failure
kurtosis run -p ./my-package

# Custom entry point file
kurtosis run --main-file deploy.star ./my-package

# Custom main function
kurtosis run --main-function-name setup ./my-package

# Don't forward ports locally
kurtosis run --no-connect ./my-package

# Show dependency graph
kurtosis run --output-graph ./my-package

# List image and package dependencies
kurtosis run --dependencies ./my-package

# Pull all dependencies locally
kurtosis run --pull --dependencies ./my-package

# Skip enclave inspect output
kurtosis run --show-enclave-inspect=false ./my-package
```

## Debug mode

```bash
# Run with debug engine images
kurtosis --debug-mode run ./my-package

# Increase CLI log verbosity
kurtosis --cli-log-level debug run ./my-package
```
