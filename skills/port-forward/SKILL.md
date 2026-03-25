---
name: port-forward
description: View and manage port mappings for Kurtosis services. Check which local ports map to service ports and troubleshoot connectivity. Use when services aren't reachable or you need to find the right port.
compatibility: Requires kurtosis CLI with a running engine and at least one enclave.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Port Forward

View and manage port mappings for Kurtosis services.

## View port info

```bash
kurtosis port print <enclave-name> <service-name> <port-id>
```

This prints the local URL for a specific port.

## Find all ports

The easiest way to see all port mappings is enclave inspect:

```bash
kurtosis enclave inspect <enclave-name>
```

Output shows mappings like:

```
rpc: 8545/tcp -> 127.0.0.1:61817
ws: 8546/tcp -> 127.0.0.1:61813
```

The right side (`127.0.0.1:61817`) is how you access the service locally.

## Port IDs

Port IDs are defined in the Starlark ServiceConfig:

```python
plan.add_service(
    name="my-service",
    config=ServiceConfig(
        image="ethereum/client-go:latest",
        ports={
            "rpc": PortSpec(number=8545),      # port ID: "rpc"
            "ws": PortSpec(number=8546),        # port ID: "ws"
            "metrics": PortSpec(number=9001),   # port ID: "metrics"
        },
    ),
)
```

## No-connect mode

If you ran with `--no-connect`, ports won't be forwarded locally:

```bash
# Ports not forwarded
kurtosis run --no-connect ./my-package

# Ports forwarded (default)
kurtosis run ./my-package
```

## Kubernetes

On Kubernetes, port forwarding goes through the gateway. If ports stop working, restart the gateway:

```bash
pkill -f "kurtosis gateway"
kurtosis gateway &
```

## Troubleshooting

When a port is unreachable, follow this diagnostic workflow:

```bash
# 1. Verify the enclave is running and check port mappings
kurtosis enclave inspect <enclave-name>

# 2. Confirm the specific service is running (not STOPPED)
kurtosis service inspect <enclave-name> <service-name>

# 3. If service is running but port fails, check service logs for errors
kurtosis service logs <enclave-name> <service-name>

# 4. On Kubernetes, verify the gateway is running
kurtosis gateway
```

| Symptom | Cause | Fix |
|---------|-------|-----|
| Port not accessible | Service not running or port not exposed | Check step 1-2 above |
| Connection refused | Service crashed or wrong interface | Check service logs (step 3) |
| No port mapping shown | Ran with `--no-connect` | Re-run without `--no-connect` |
| K8s ports unreachable | Gateway not running | Restart gateway (step 4) |
