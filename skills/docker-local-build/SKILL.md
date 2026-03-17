---
name: docker-local-build
description: Build and test Kurtosis from source on local Docker. Compiles all components (engine, core, files-artifacts-expander), builds Docker images, installs the CLI, and restarts the engine. Use when developing Kurtosis and testing changes locally with Docker.
compatibility: Requires Go 1.23+, Docker, and docker buildx.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Docker Local Build

Build all Kurtosis components from source and test them locally on Docker.

## Overview

This skill builds the full Kurtosis stack for local Docker testing:
- **Engine** — Docker image `kurtosistech/engine`
- **Core (APIC)** — Docker image `kurtosistech/core`
- **Files Artifacts Expander** — Docker image `kurtosistech/files-artifacts-expander`
- **CLI** — binary at `.tmp/kurtosis`, installed to `/usr/local/bin/kurtosis`

All images are tagged with the current git short SHA (+ `-dirty` if uncommitted changes exist) via `scripts/get-docker-tag.sh`. The same tag is compiled into the CLI as `KurtosisVersion`.

## Quick build (recommended)

Run the wrapper script from the repo root. It handles version generation, image builds, CLI compilation, installation, and engine reset:

```bash
./scripts/local-build.sh
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `INSTALL_PATH` | `/usr/local/bin/kurtosis` | Where to install the CLI binary |
| `BUILD_IMAGES` | `true` | Set to `false` to skip Docker image builds (CLI only) |
| `DEBUG_IMAGE` | `false` | Build debug images with delve debugger |
| `PODMAN_MODE` | `false` | Use podman instead of docker |
| `RESET_KURTOSIS_AFTER_BUILD` | `true` | Stop engine and clean containers after build |

### Examples

```bash
# Full build + install + reset
./scripts/local-build.sh

# CLI only (skip image builds)
BUILD_IMAGES=false ./scripts/local-build.sh

# Build with debug images
DEBUG_IMAGE=true ./scripts/local-build.sh

# Install to a custom path
INSTALL_PATH=/tmp/kurtosis ./scripts/local-build.sh

# Skip engine reset (keep running enclaves)
RESET_KURTOSIS_AFTER_BUILD=false ./scripts/local-build.sh
```

## Manual build (step by step)

### 1. Generate version constants

```bash
./scripts/generate-kurtosis-version.sh
```

This writes the git-based version tag into `kurtosis_version/kurtosis_version.go`.

### 2. Build Docker images

The existing build scripts handle binary compilation, unit tests, and Docker image building:

```bash
# Build engine image (compiles binary, runs tests, builds Docker image)
./engine/scripts/build.sh

# Build core + files-artifacts-expander images
./core/scripts/build.sh
```

These scripts:
- Detect the host architecture (amd64/arm64)
- Cross-compile the Go binary for linux
- Run unit tests
- Build a Docker image tagged `kurtosistech/<component>:<git-sha>`
- Use `scripts/docker-image-builder.sh` which creates a buildx builder, builds for the local platform with `--load`, and cleans up

### 3. Build the CLI

```bash
go build -o .tmp/kurtosis ./cli/cli/
```

### 4. Install

```bash
cp .tmp/kurtosis /usr/local/bin/kurtosis
```

### 5. Reset the engine

```bash
kurtosis engine stop || true
docker ps -aq --filter "label=com.kurtosistech.app-id=kurtosis" | xargs -r docker rm -f
docker ps -aq --filter "name=kurtosis-engine" | xargs -r docker rm -f
docker ps -aq --filter "name=kurtosis-reverse-proxy" | xargs -r docker rm -f
docker ps -aq --filter "name=kurtosis-logs-collector" | xargs -r docker rm -f
docker ps -aq --filter "name=kurtosis-api--" | xargs -r docker rm -f
```

### 6. Start and test

```bash
kurtosis engine start
kurtosis run github.com/ethpandaops/ethereum-package
kurtosis clean -a
```

## Key files

| Component | Source | Dockerfile | Image |
|-----------|--------|------------|-------|
| Engine | `engine/server/engine/main.go` | `engine/server/Dockerfile` | `kurtosistech/engine` |
| Core (APIC) | `core/server/api_container/main.go` | `core/server/Dockerfile` | `kurtosistech/core` |
| Files Artifacts Expander | `core/files_artifacts_expander/main.go` | `core/files_artifacts_expander/Dockerfile` | `kurtosistech/files-artifacts-expander` |
| CLI | `cli/cli/main.go` | N/A | N/A |
| Version | `kurtosis_version/kurtosis_version.go` | N/A | N/A |
| Version generator | `scripts/generate-kurtosis-version.sh` | N/A | N/A |
| Docker tag | `scripts/get-docker-tag.sh` | N/A | N/A |
| Image builder | `scripts/docker-image-builder.sh` | N/A | N/A |

## Iterating

After making code changes:

1. Re-run `./scripts/local-build.sh` (or set `BUILD_IMAGES=false` if only CLI changed)
2. The version tag changes automatically with each commit (SHA-based)
3. If you have uncommitted changes, the tag gets a `-dirty` suffix

## Common issues

| Symptom | Fix |
|---------|-----|
| `go build` fails with import errors | Run `go mod tidy` in the failing module directory |
| Docker build fails with `.dockerignore` error | Ensure `.dockerignore` exists in the component's server directory |
| Engine starts but uses old images | The tag is SHA-based; make sure the built image tag matches `kurtosis version` output |
| `webapp directory not found` warning | Normal when building without enclave-manager; an empty placeholder is created automatically |
| buildx builder conflict | Remove stale builders: `docker buildx rm kurtosis-docker-builder` and `docker context rm kurtosis-docker-builder-context` |
| Version mismatch between CLI and engine | Rebuild everything from the same commit; version is compiled into the CLI binary |
