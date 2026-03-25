---
name: gateway
description: Start and manage the Kurtosis gateway for Kubernetes. The gateway forwards local ports to the Kurtosis engine and services running in a k8s cluster. Required when using Kurtosis with Kubernetes. Use when kurtosis engine status shows nothing on k8s or services aren't reachable.
compatibility: Requires kurtosis CLI with Kubernetes cluster access.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Gateway

The Kurtosis gateway creates local port forwards to the engine and services running in a Kubernetes cluster.

## When you need it

The gateway is **required** when running Kurtosis on Kubernetes. Without it, the local CLI cannot reach the engine pod in the cluster.

Not needed when using Docker backend.

## Start the gateway

```bash
# Run in the background
kurtosis gateway &

# Or in a separate terminal
kurtosis gateway
```

## Verify it's working

```bash
kurtosis engine status
```

If this returns engine info, the gateway is working. If it says "No Kurtosis engine is running" but you know the engine pod is up, the gateway isn't running.

## Stop the gateway

```bash
pkill -f "kurtosis gateway"
```

## How it works

The gateway:
1. Finds the engine pod in the `kurtosis-engine-*` namespace
2. Creates a local port forward to the engine's gRPC port
3. When services are accessed, creates additional port forwards to service pods
4. Port mappings shown in `kurtosis enclave inspect` point to localhost via the gateway

## Common issues

| Symptom | Fix |
|---------|-----|
| `No engine running` but engine pod is up | Start the gateway: `kurtosis gateway &` |
| Gateway crashes or disconnects | Restart: `pkill -f "kurtosis gateway"; kurtosis gateway &` |
| Port conflicts | Kill old gateway first: `pkill -f "kurtosis gateway"` |
| Services unreachable after gateway restart | Re-inspect enclave for new port mappings |
