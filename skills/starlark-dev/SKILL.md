---
name: starlark-dev
description: Develop and debug Kurtosis Starlark packages. Create packages from scratch, understand the plan-based execution model, use print() debugging, handle future references, and test packages locally. Use when writing or troubleshooting .star files.
compatibility: Requires kurtosis CLI.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Starlark Dev

Create, debug, and test Kurtosis Starlark packages.

## Package structure

A minimal Kurtosis package needs two files:

```
my-package/
  kurtosis.yml    # Package metadata
  main.star       # Entry point
```

### kurtosis.yml

```yaml
name: github.com/your-org/my-package
```

### main.star

```python
def run(plan, args):
    plan.add_service(
        name="my-service",
        config=ServiceConfig(
            image="nginx:latest",
            ports={
                "http": PortSpec(number=80, transport_protocol="TCP"),
            },
        ),
    )
```

## Running packages

```bash
# Run a local package
kurtosis run ./my-package

# Run with parameters
kurtosis run ./my-package '{"param1": "value1"}'

# Run a remote package from GitHub
kurtosis run github.com/ethpandaops/ethereum-package

# Run with a custom config file
kurtosis run github.com/ethpandaops/ethereum-package --args-file config.yaml

# Dry run (plan only, no execution)
kurtosis run ./my-package --dry-run
```

## Execution model

Kurtosis Starlark executes in two phases:

1. **Planning phase** — Your code runs and builds a plan of actions. `add_service()`, `exec()`, etc. don't execute immediately — they return future references.
2. **Execution phase** — The plan is executed in order. Future references are resolved to actual values.

This means you **cannot** use the return value of `plan.exec()` in Python-level logic like `if/else` during the planning phase. Use `plan.verify()` or `plan.assert()` instead.

```python
# WRONG: result is a future reference, not a real value during planning
result = plan.exec(service_name="my-service", recipe=ExecRecipe(command=["echo", "hello"]))
if result["output"] == "hello":  # This won't work as expected
    plan.print("matched")

# RIGHT: use plan.verify for conditional checks
result = plan.exec(service_name="my-service", recipe=ExecRecipe(command=["echo", "hello"]))
plan.verify(result["exit_code"], "==", 0)
```

## Debugging with print

```python
def run(plan, args):
    plan.print("Args received: {}".format(args))

    service = plan.add_service(
        name="my-service",
        config=ServiceConfig(image="nginx:latest"),
    )
    plan.print("Service IP: {}".format(service.ip_address))
    plan.print("Service hostname: {}".format(service.hostname))
```

## Common patterns

### Wait for service readiness

```python
plan.wait(
    service_name="my-service",
    recipe=GetHttpRequestRecipe(port_id="http", endpoint="/health"),
    field="code",
    assertion="==",
    target_value=200,
    timeout="60s",
)
```

### Execute commands in a service

```python
result = plan.exec(
    service_name="my-service",
    recipe=ExecRecipe(command=["cat", "/etc/hostname"]),
)
plan.verify(result["exit_code"], "==", 0)
plan.print("Hostname: {}".format(result["output"]))
```

### Upload files

```python
config_template = read_file("./templates/config.toml")
artifact = plan.render_templates(
    name="my-config",
    config={
        "config.toml": struct(
            template=config_template,
            data={"key": "value"},
        ),
    },
)

plan.add_service(
    name="my-service",
    config=ServiceConfig(
        image="my-image:latest",
        files={"/etc/myapp": artifact},
    ),
)
```

### Import from other packages

```python
dependency = import_module("github.com/org/other-package/lib.star")

def run(plan, args):
    dependency.some_function(plan)
```

## Testing

```bash
# Run and check output
kurtosis run ./my-package

# Inspect the created enclave
kurtosis enclave inspect <enclave-name>

# Check service logs
kurtosis service logs <enclave-name> <service-name>

# Shell into a service to verify state
kurtosis service shell <enclave-name> <service-name>

# Clean up after testing
kurtosis clean -a
```

## Common errors

| Error | Cause | Fix |
|-------|-------|-----|
| `cannot use future reference in if` | Using plan result in Python logic | Use `plan.verify()` or `plan.assert()` |
| `service not found` | Service name typo or not yet created | Check `plan.add_service()` name matches |
| `port not found` | Port ID mismatch | Ensure `port_id` in recipes matches `ports` dict key |
| `image pull failed` | Image doesn't exist or no auth | Verify image tag, check `docker pull` manually |
| `kurtosis.yml not found` | Running from wrong directory | Run from package root containing `kurtosis.yml` |
