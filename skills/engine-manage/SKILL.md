---
name: engine-manage
description: Manage the Kurtosis engine server. Start, stop, restart the engine, check status, and view engine logs. Covers both Docker and Kubernetes engine backends. Use when the engine won't start, needs restarting, or you need to check engine health.
compatibility: Requires kurtosis CLI. Docker or Kubernetes cluster access depending on backend.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Engine Manage

Manage the Kurtosis engine server lifecycle.

## Status

```bash
kurtosis engine status
```

Shows whether the engine is running and its version.

## Start

```bash
# Start with default settings
kurtosis engine start

# Start with debug images
kurtosis --debug-mode engine start
```

On Kubernetes, the engine runs as a pod in a `kurtosis-engine-*` namespace. You also need to run `kurtosis gateway` to access it from your local machine.

## Stop

```bash
kurtosis engine stop
```

## Restart

```bash
kurtosis engine restart
```

Equivalent to stop + start. Useful after changing cluster settings or upgrading.

## View logs

```bash
kurtosis engine logs
```

Dumps engine server logs. Useful for diagnosing startup failures or API errors.

## Docker vs Kubernetes

```bash
# Check which backend is active
kurtosis cluster get

# Switch to Docker
kurtosis cluster set docker
kurtosis engine restart

# Switch to Kubernetes
kurtosis cluster set kubernetes
kurtosis engine restart
kurtosis gateway  # Required for k8s
```

## Engine on Kubernetes

When running on Kubernetes:

```bash
# The engine runs in its own namespace
kubectl get ns | grep kurtosis-engine

# Check engine pod
kubectl get pods -n <engine-namespace>

# View engine logs directly
kubectl logs <engine-pod> -n <engine-namespace>

# Start the gateway (required for local CLI to reach k8s engine)
kurtosis gateway &
```

## Common issues

| Symptom | Fix |
|---------|-----|
| `No Kurtosis engine is running` | Run `kurtosis engine start` |
| Engine starts but `engine status` shows nothing (k8s) | Start the gateway: `kurtosis gateway` |
| Version mismatch warning | `kurtosis engine restart` to match CLI version |
| Engine start hangs (k8s) | Check pods: `kubectl get pods -A \| grep kurtosis` |
| Old engine blocking new start | `kurtosis engine stop` then clean namespaces |
