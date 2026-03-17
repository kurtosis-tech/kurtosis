---
name: enclave-inspect
description: Inspect and manage Kurtosis enclaves. List enclaves, view services and ports, examine file artifacts, dump enclave state for debugging, and clean up. Use when you need to understand what's running inside an enclave or export its state.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Enclave Inspect

Inspect and manage Kurtosis enclaves and their contents.

## List enclaves

```bash
kurtosis enclave ls
```

Output shows enclave name, UUID, status (RUNNING/STOPPED), and creation time.

## Inspect an enclave

```bash
kurtosis enclave inspect <enclave-name>
```

This shows:
- **File Artifacts** — uploaded files and rendered templates
- **User Services** — running containers with their ports and status

## Services

### View service details

```bash
# Full enclave inspection (shows all services with ports)
kurtosis enclave inspect <enclave-name>

# View logs for a service
kurtosis service logs <enclave-name> <service-name>

# Follow logs in real time
kurtosis service logs <enclave-name> <service-name> -f

# Shell into a service
kurtosis service shell <enclave-name> <service-name>

# Run a command in a service
kurtosis service exec <enclave-name> <service-name> -- <command>
```

### Access service ports

The `inspect` output shows port mappings like:

```
http: 8545/tcp -> http://127.0.0.1:61817
```

This means port 8545 inside the container is mapped to localhost:61817. On Kubernetes with gateway running, these are forwarded through the gateway.

## File artifacts

File artifacts are named blobs of files stored in the enclave.

```bash
# List artifacts (shown in enclave inspect)
kurtosis enclave inspect <enclave-name>

# Download an artifact to local disk
kurtosis files download <enclave-name> <artifact-name> /tmp/output-dir

# Upload a local file as an artifact
kurtosis files upload <enclave-name> /path/to/local/file
```

## Dump enclave state

Export the full enclave state for offline debugging:

```bash
kurtosis enclave dump <enclave-name> /tmp/enclave-dump
```

This exports:
- Service logs
- Service configurations
- File artifacts
- Enclave metadata

Useful for sharing with others to debug issues.

## Enclave lifecycle

```bash
# Create an enclave (usually done by kurtosis run)
kurtosis enclave add <enclave-name>

# Stop an enclave (preserves state)
kurtosis enclave stop <enclave-name>

# Remove a specific enclave
kurtosis enclave rm <enclave-name>

# Remove all enclaves
kurtosis clean -a
```

## Kubernetes-specific

On Kubernetes, each enclave is a namespace prefixed with `kt-`:

```bash
# See enclave namespaces
kubectl get ns | grep kt-

# See pods in an enclave
kubectl get pods -n kt-<enclave-name>

# Describe a service pod
kubectl describe pod <pod-name> -n kt-<enclave-name>

# View pod logs directly
kubectl logs <pod-name> -n kt-<enclave-name>
```
