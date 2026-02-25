---
name: k8s-clean-cluster
description: Force-clean all Kurtosis resources from a Kubernetes cluster when kurtosis clean hangs or fails. Removes all kurtosis namespaces, pods, daemonsets, cluster roles, and cluster role bindings. Use when kurtosis clean -a hangs or leaves behind orphaned resources.
compatibility: Requires kubectl with cluster access.
metadata:
  author: ethpandaops
  version: "1.0"
---

# K8s Clean Cluster

Force-clean all Kurtosis resources from a Kubernetes cluster when the normal `kurtosis clean -a` command hangs or fails.

## When to use

- `kurtosis clean -a` hangs for more than a few minutes
- Orphaned kurtosis namespaces remain after a failed clean
- `remove-dir-pod-*` pods are stuck in Pending state
- Engine start fails because old resources exist

## Steps

### 1. Kill any running kurtosis processes

```bash
pkill -f "kurtosis gateway" 2>/dev/null
pkill -f "kurtosis clean" 2>/dev/null
```

### 2. Stop the engine gracefully (if possible)

```bash
kurtosis engine stop || true
```

### 3. Delete all kurtosis namespaces

```bash
# List them first
kubectl get ns | grep kurtosis

# Delete all kurtosis namespaces (engine, enclaves, logs)
kubectl get ns | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete ns --force --grace-period=0
```

### 4. Clean up cluster-scoped resources

```bash
# Delete kurtosis cluster roles
kubectl get clusterrole | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete clusterrole

# Delete kurtosis cluster role bindings
kubectl get clusterrolebinding | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete clusterrolebinding
```

### 5. Clean up stuck pods

```bash
# Force-delete any remaining kurtosis pods
kubectl get pods -A | grep kurtosis | awk '{print $2 " -n " $1}' | xargs -L1 kubectl delete pod --force --grace-period=0

# Clean up evicted pods
kubectl get pods -A | grep Evicted | awk '{print $2 " -n " $1}' | xargs -L1 kubectl delete pod --force --grace-period=0
```

### 6. Verify clean state

```bash
kubectl get ns | grep kurtosis
kubectl get pods -A | grep kurtosis
kubectl get ds -A | grep kurtosis
```

All three commands should return empty results.

### 7. Restart

```bash
kurtosis engine start
kurtosis gateway &
```

## Why clean hangs

The most common cause is the fluentbit logs collector `Clean` method which:
1. Evicts all DaemonSet pods by adding a non-existent node selector
2. Waits for each pod to terminate (up to 5 min per pod, sequentially)
3. Creates `remove-dir-pod` cleanup pods targeted at each node
4. Cleanup pods on tainted/unhealthy nodes get stuck in Pending

The fix in the codebase makes these operations best-effort with timeouts and detects unschedulable pods early, but if running an unfixed version, manual cleanup is needed.
