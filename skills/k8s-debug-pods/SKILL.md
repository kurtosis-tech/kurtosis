---
name: k8s-debug-pods
description: Debug Kurtosis pods on Kubernetes. Diagnose why pods are Pending, CrashLoopBackOff, ImagePullBackOff, or Evicted. Check node taints, tolerations, resource pressure, and pod events. Use when kurtosis engine start fails or pods aren't coming online.
compatibility: Requires kubectl with cluster access.
metadata:
  author: ethpandaops
  version: "1.0"
---

# K8s Debug Pods

Diagnose and fix issues with Kurtosis pods on Kubernetes.

## Quick triage

```bash
# See all kurtosis-related pods across namespaces
kubectl get pods -A | grep kurtosis

# Check for problem pods (not Running)
kubectl get pods -A | grep kurtosis | grep -v Running

# Get events for a specific pod
kubectl describe pod <POD_NAME> -n <NAMESPACE> | tail -30
```

## Common pod states and fixes

### Pending — Unschedulable

The pod can't be scheduled because of node taints, resource pressure, or affinity rules.

```bash
# Check node taints
kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints

# Check node conditions (DiskPressure, MemoryPressure, etc.)
kubectl get nodes -o custom-columns=NAME:.metadata.name,CONDITIONS:.status.conditions[*].type
```

**Fix**: Add tolerations to the kurtosis config at `~/Library/Application Support/kurtosis/kurtosis-config.yml` or fix the node condition.

### ImagePullBackOff

The image tag doesn't exist on the registry.

```bash
# Check which image is failing
kubectl describe pod <POD_NAME> -n <NAMESPACE> | grep -A5 "Image:"

# Verify image exists on Docker Hub
docker manifest inspect <IMAGE>:<TAG>
```

**Fix**: Push the correct image tag, or fix the image reference in the code.

### CrashLoopBackOff

The container starts but crashes immediately.

```bash
# Check container logs
kubectl logs <POD_NAME> -n <NAMESPACE>
kubectl logs <POD_NAME> -n <NAMESPACE> --previous
```

### Evicted

The node evicted the pod due to resource pressure.

```bash
# Check which nodes have pressure
kubectl get nodes -o custom-columns=NAME:.metadata.name,STATUS:.status.conditions[-1].type

# Clean up evicted pods
kubectl get pods -A | grep Evicted | awk '{print $2 " -n " $1}' | xargs -L1 kubectl delete pod
```

## Kurtosis-specific pod types

| Pod pattern | Component | Image source |
|-------------|-----------|-------------|
| `kurtosis-engine-*` | Engine server | `engine/server/Dockerfile` |
| `kurtosis-api` (in `kt-*` namespaces) | API Container (APIC) | `core/server/Dockerfile` |
| `kurtosis-logs-collector-*` | Fluentbit DaemonSet | Pulled from registry |
| `kurtosis-logs-aggregator-*` | Vector deployment | Pulled from registry |
| `remove-dir-pod-*` | Fluentbit cleanup pods | busybox |
| `files-artifact-expander` (init container) | Files artifacts | `core/files_artifacts_expander/Dockerfile` |

## Engine start failures

If `kurtosis engine start` fails:

1. Check if old kurtosis namespaces exist: `kubectl get ns | grep kurtosis`
2. Delete them: `kubectl get ns | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete ns`
3. Retry engine start

## Logs collector issues

The logs collector is a DaemonSet that runs on every node. If some nodes are unhealthy:

```bash
# Check DaemonSet status
kubectl get ds -A | grep kurtosis

# See which pods are not running
kubectl get pods -A | grep logs-collector | grep -v Running
```

Nodes with DiskPressure or other taints may not schedule collector pods — this is expected and the engine should start with a warning about partially degraded collection.
