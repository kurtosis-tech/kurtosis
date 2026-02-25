---
name: cluster-manage
description: Manage Kurtosis cluster settings. Switch between Docker and Kubernetes backends, list available clusters, and configure which cluster Kurtosis uses. Use when you need to change where Kurtosis runs enclaves.
compatibility: Requires kurtosis CLI. Kubernetes requires kubectl with cluster access.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Cluster Manage

Switch between Docker and Kubernetes backends for Kurtosis.

## Check current cluster

```bash
kurtosis cluster get
```

Returns `docker` or `kubernetes`.

## List available clusters

```bash
kurtosis cluster ls
```

## Switch cluster

```bash
# Switch to Docker
kurtosis cluster set docker

# Switch to Kubernetes (uses current kubectl context)
kurtosis cluster set kubernetes
```

After switching, restart the engine:

```bash
kurtosis engine restart
```

## Kubernetes setup

When using Kubernetes:

1. Ensure `kubectl` is configured and can reach your cluster:
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

2. Switch Kurtosis to Kubernetes:
   ```bash
   kurtosis cluster set kubernetes
   kurtosis engine start
   ```

3. Start the gateway (required for local CLI to reach the k8s-based engine):
   ```bash
   kurtosis gateway &
   ```

4. Verify:
   ```bash
   kurtosis engine status
   ```

## Config file

The cluster setting is stored in the Kurtosis config file:

```bash
kurtosis config path
```

Typically at `~/Library/Application Support/kurtosis/kurtosis-config.yml` on macOS.
