---
name: docker-debug
description: Debug Kurtosis running on local Docker. Inspect engine, API container, and service logs. Diagnose container crashes, port conflicts, and networking issues. Use when kurtosis commands fail or services aren't reachable on Docker.
compatibility: Requires Docker and kurtosis CLI.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Docker Debug

Diagnose and fix issues with Kurtosis running on a local Docker engine.

## Quick triage

```bash
# Check engine is running
kurtosis engine status

# List all kurtosis containers
docker ps -a --filter "label=app.kubernetes.io/managed-by=kurtosis" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# If no label filter works, grep for kurtosis
docker ps -a | grep kurtosis
```

## Engine issues

### Engine won't start

```bash
# Check if old engine is still running
docker ps -a | grep kurtosis-engine

# Check engine logs
docker logs $(docker ps -aq --filter "name=kurtosis-engine") 2>&1 | tail -50

# Nuclear option: stop and remove all kurtosis containers
kurtosis engine stop
docker ps -a | grep kurtosis | awk '{print $1}' | xargs -r docker rm -f
kurtosis engine start
```

### Engine version mismatch

```bash
# Check CLI version
kurtosis version

# Check running engine version
kurtosis engine status

# Force restart with matching version
kurtosis engine restart
```

## Enclave / API container issues

The API container (core/APIC) runs inside each enclave and manages services.

```bash
# List enclaves
kurtosis enclave ls

# Find the APIC container for an enclave
docker ps -a | grep "kurtosis-api"

# View APIC logs (most useful for debugging enclave creation failures)
docker logs $(docker ps -aq --filter "name=kurtosis-api") 2>&1 | tail -100
```

## Service debugging

```bash
# List services in an enclave
kurtosis enclave inspect <enclave-name>

# View service logs
kurtosis service logs <enclave-name> <service-name>

# Follow logs in real time
kurtosis service logs <enclave-name> <service-name> -f

# Shell into a running service
kurtosis service shell <enclave-name> <service-name>

# Execute a command in a service
kurtosis service exec <enclave-name> <service-name> -- <command>
```

## Port and networking issues

```bash
# Check mapped ports for a service
kurtosis enclave inspect <enclave-name>

# Verify port is actually listening inside the container
kurtosis service exec <enclave-name> <service-name> -- netstat -tlnp

# Test connectivity between services (from inside a service)
kurtosis service exec <enclave-name> <service-name> -- wget -qO- http://<other-service>:<port>/endpoint
```

## File artifacts

```bash
# List file artifacts in an enclave
kurtosis enclave inspect <enclave-name>

# Download a file artifact for inspection
kurtosis files download <enclave-name> <artifact-name> /tmp/artifact-output
```

## Common problems

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `engine not running` | Engine crashed or was stopped | `kurtosis engine start` |
| Port conflict on start | Old container holding the port | `docker ps -a \| grep kurtosis \| awk '{print $1}' \| xargs docker rm -f` |
| Service unreachable | Wrong port or service not ready | Check `kurtosis enclave inspect` for mapped ports |
| `image not found` | Image not pulled or tag wrong | Check image name in Starlark, try `docker pull <image>` |
| Enclave creation hangs | APIC crash or image pull issue | Check APIC logs: `docker logs` on the kurtosis-api container |

## Cleanup

```bash
# Remove a specific enclave
kurtosis enclave rm <enclave-name>

# Remove all enclaves and clean up
kurtosis clean -a

# Full nuclear clean (if kurtosis clean fails)
docker ps -a | grep kurtosis | awk '{print $1}' | xargs -r docker rm -f
docker network ls | grep kurtosis | awk '{print $1}' | xargs -r docker network rm
kurtosis engine start
```
