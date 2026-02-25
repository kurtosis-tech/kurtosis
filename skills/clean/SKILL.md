---
name: clean
description: Clean up Kurtosis enclaves and artifacts. Remove stopped enclaves, running enclaves with -a flag, and stopped engine containers. Use when you need to free up resources or start fresh.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Clean

Remove Kurtosis enclaves and leftover artifacts.

## Basic clean

Removes only **stopped** enclaves and stopped engine containers:

```bash
kurtosis clean
```

## Clean everything

Removes **all** enclaves (including running ones):

```bash
kurtosis clean -a
```

## Selective removal

To remove a specific enclave without touching others:

```bash
# Stop an enclave
kurtosis enclave stop <enclave-name>

# Remove a specific enclave
kurtosis enclave rm <enclave-name>
```

## When clean hangs

On Kubernetes, `kurtosis clean -a` can hang if the logs collector cleanup tries to create pods on tainted/unhealthy nodes. See the `k8s-clean-cluster` skill for manual cleanup steps.

On Docker, if clean hangs:

```bash
# Kill the hanging process
pkill -f "kurtosis clean"

# Manually remove containers
docker ps -a | grep kurtosis | awk '{print $1}' | xargs -r docker rm -f

# Remove networks
docker network ls | grep kurtosis | awk '{print $1}' | xargs -r docker network rm

# Restart engine
kurtosis engine start
```
