---
name: service-manage
description: Manage services in Kurtosis enclaves. Add, inspect, stop, start, remove, update services. View logs, shell into containers, and execute commands. Use when you need to interact with running services.
compatibility: Requires kurtosis CLI with a running engine and at least one enclave.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Service Manage

Manage services running inside Kurtosis enclaves.

## List services

```bash
# Services are shown in enclave inspect output
kurtosis enclave inspect <enclave-name>
```

## View logs

```bash
# View logs
kurtosis service logs <enclave-name> <service-name>

# Follow logs in real time
kurtosis service logs <enclave-name> <service-name> -f

# Show all logs (not just recent)
kurtosis service logs <enclave-name> <service-name> -a
```

## Shell and exec

```bash
# Get an interactive shell
kurtosis service shell <enclave-name> <service-name>

# Execute a single command
kurtosis service exec <enclave-name> <service-name> -- ls -la /data

# Execute with pipes (wrap in sh -c)
kurtosis service exec <enclave-name> <service-name> -- sh -c "cat /etc/hosts | grep localhost"
```

## Inspect a service

```bash
kurtosis service inspect <enclave-name> <service-name>
```

Shows detailed info including ports, status, and container ID.

## Stop and start

```bash
# Stop a service (keeps it in the enclave, just stops the container)
kurtosis service stop <enclave-name> <service-name>

# Restart a stopped service
kurtosis service start <enclave-name> <service-name>
```

## Remove a service

```bash
kurtosis service rm <enclave-name> <service-name>
```

## Add a service manually

```bash
kurtosis service add <enclave-name> <service-name> <image>
```

## Update a service

```bash
kurtosis service update <enclave-name> <service-name>
```

## Common patterns

### Check if a service is healthy

```bash
# HTTP health check
kurtosis service exec <enclave-name> <service-name> -- wget -qO- http://localhost:8080/health

# Check process is running
kurtosis service exec <enclave-name> <service-name> -- ps aux

# Check listening ports
kurtosis service exec <enclave-name> <service-name> -- netstat -tlnp
```

### Debug a crashing service

```bash
# Check recent logs
kurtosis service logs <enclave-name> <service-name>

# Check all logs from the start
kurtosis service logs <enclave-name> <service-name> -a

# Inspect for error status
kurtosis service inspect <enclave-name> <service-name>
```

### Copy data between services

Use file artifacts in Starlark:

```python
# Store files from one service
artifact = plan.store_service_files(service_name="source-svc", src="/data/output", name="shared-data")

# Mount in another service
plan.add_service(name="dest-svc", config=ServiceConfig(
    image="my-image",
    files={"/input": artifact},
))
```
