---
name: k8s-dev-deploy
description: Build, push, and deploy Kurtosis dev images to a Kubernetes cluster without creating a release. Rebuilds engine, core, and files-artifacts-expander as multi-arch Docker images with a unique tag, pushes to the logged-in user's Docker Hub, and restarts the engine. Use when testing local code changes on a k8s cluster.
compatibility: Requires docker buildx, go, kubectl, and Docker Hub login.
metadata:
  author: ethpandaops
  version: "1.0"
---

# K8s Dev Deploy

Build and deploy Kurtosis from source to a Kubernetes cluster for testing without making a release.

## Overview

Three container images must be built and pushed:
- `engine` — from `engine/server/Dockerfile`
- `core` (APIC) — from `core/server/Dockerfile`
- `files-artifacts-expander` — from `core/files_artifacts_expander/Dockerfile`

The CLI binary is also rebuilt locally.

All images share the same tag, which is set in `kurtosis_version/kurtosis_version.go` as `KurtosisVersion`. This version is compiled into the engine binary and used at runtime to pull the core and files-artifacts-expander images.

## Steps

### 1. Determine Docker Hub username and generate a unique tag

```bash
DOCKER_USER=$(docker info 2>/dev/null | grep Username | awk '{print $2}')
TAG="$(git rev-parse --short HEAD)-$(date +%s)"
```

### 2. Update image references to use the user's Docker Hub

Edit these three constants to replace `kurtosistech` with `$DOCKER_USER`:

| File | Constant | Default Value |
|------|----------|--------------|
| `engine/launcher/engine_server_launcher/engine_server_launcher.go` | `containerImage` | `kurtosistech/engine` |
| `core/launcher/api_container_launcher/api_container_launcher.go` | `containerImage` | `kurtosistech/core` |
| `core/server/api_container/server/startosis_engine/kurtosis_types/service_config/service_config.go` | `filesArtifactsExpanderImage` | `kurtosistech/files-artifacts-expander` |

### 3. Set the version tag

Edit `kurtosis_version/kurtosis_version.go` and set `KurtosisVersion` to the generated `$TAG`.

**IMPORTANT**: Always use a unique tag (include a timestamp) to avoid Kubernetes node image caching issues. Nodes with `imagePullPolicy: IfNotPresent` won't re-pull an image with the same tag.

### 4. Build all binaries (parallel)

```bash
# Engine
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o engine/server/build/kurtosis-engine.arm64 engine/server/engine/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o engine/server/build/kurtosis-engine.amd64 engine/server/engine/main.go

# Core (APIC)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o core/server/build/api-container.arm64 core/server/api_container/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o core/server/build/api-container.amd64 core/server/api_container/main.go

# Files Artifacts Expander
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o core/files_artifacts_expander/build/files-artifacts-expander.arm64 core/files_artifacts_expander/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o core/files_artifacts_expander/build/files-artifacts-expander.amd64 core/files_artifacts_expander/main.go

# CLI (local use)
go build -o /tmp/kurtosis ./cli/cli/
```

### 5. Push multi-arch Docker images (parallel)

Ensure a buildx builder exists (create with `docker buildx create --name kurtosis-builder --use` if needed).

```bash
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/engine:$TAG --push engine/server/
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/core:$TAG --push core/server/
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/files-artifacts-expander:$TAG --push core/files_artifacts_expander/
```

**IMPORTANT**: Always use `--no-cache` to prevent buildx from caching old binaries.

### 6. Deploy to the cluster

```bash
/tmp/kurtosis engine stop
# Clean any leftover namespaces
kubectl get ns | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete ns
/tmp/kurtosis engine start
```

### 7. Start gateway and test

```bash
/tmp/kurtosis gateway &
/tmp/kurtosis run github.com/ethpandaops/ethereum-package
/tmp/kurtosis clean -a
```

## Iterating

When making further code changes, repeat from step 3 with a **new unique tag** each time. Never reuse a tag — k8s nodes cache images and won't pull updates under the same tag.

## Common issues

- **ImagePullBackOff**: The tag doesn't exist on Docker Hub. Verify the push succeeded.
- **Old code running despite push**: The k8s node has the image cached under the same tag. Use a new timestamp-based tag.
- **Engine start hangs**: Logs collector DaemonSet pods failing on tainted/unhealthy nodes. Check `kubectl get pods -A | grep kurtosis`.
- **Clean hangs**: The fluentbit clean process tries to create cleanup pods on tainted nodes. The best-effort fixes should handle this.
