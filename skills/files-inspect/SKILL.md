---
name: files-inspect
description: Inspect, download, upload, and debug Kurtosis file artifacts. View artifacts in an enclave, download them locally for inspection, upload local files, and troubleshoot file mounting issues. Use when services can't find expected files or configs are wrong.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Files Inspect

Work with Kurtosis file artifacts — the mechanism for passing files between services and into containers.

## What are file artifacts?

File artifacts are named collections of files stored in an enclave. They're created by:
- `plan.upload_files()` — upload local files from a package
- `plan.render_templates()` — render Go templates with data
- `plan.store_service_files()` — copy files from a running service
- `plan.run_sh()` / `plan.run_python()` — store output files from scripts

Services mount artifacts via the `files` parameter in `ServiceConfig`.

## List artifacts in an enclave

```bash
kurtosis enclave inspect <enclave-name>
```

The "Files Artifacts" section shows each artifact's UUID and name:

```
========================================= Files Artifacts =========================================
UUID           Name
4a0563e5a391   1-lighthouse-geth-0-127
f49b81f30a8f   el_cl_genesis_data
88c3f17013f3   jwt_file
```

## Download artifacts

```bash
# Download to a local directory
kurtosis files download <enclave-name> <artifact-name> /tmp/artifact-output

# Example: inspect genesis data
kurtosis files download <enclave-name> el_cl_genesis_data /tmp/genesis
ls -la /tmp/genesis/
cat /tmp/genesis/config.yaml
```

## Upload files

```bash
# Upload a local file or directory as an artifact
kurtosis files upload <enclave-name> /path/to/local/file-or-dir
```

The command returns the artifact name and UUID for use in subsequent service configs.

## Inspect files inside a running service

When you need to verify files were mounted correctly:

```bash
# Shell into the service and browse
kurtosis service shell <enclave-name> <service-name>
ls -la /mounted/path/
cat /mounted/path/config.yaml

# Or run a one-off command
kurtosis service exec <enclave-name> <service-name> -- ls -la /mounted/path/
kurtosis service exec <enclave-name> <service-name> -- cat /mounted/path/config.yaml
```

## Starlark file patterns

### Upload files from package

```python
artifact = plan.upload_files(src="./static_files/config.yaml", name="my-config")

plan.add_service(
    name="my-service",
    config=ServiceConfig(
        image="my-image:latest",
        files={"/etc/myapp": artifact},
    ),
)
```

### Render templates with variables

```python
template = read_file("./templates/config.toml.tmpl")

artifact = plan.render_templates(
    name="rendered-config",
    config={
        "config.toml": struct(
            template=template,
            data={"port": 8080, "host": "0.0.0.0"},
        ),
    },
)
```

Template syntax uses Go templates:

```toml
# config.toml.tmpl
host = "{{.host}}"
port = {{.port}}
```

### Copy files from a running service

```python
artifact = plan.store_service_files(
    service_name="my-service",
    src="/data/output",
    name="service-output",
)
```

### Store output from a shell command

```python
result = plan.run_sh(
    run="echo 'hello' > /tmp/output.txt && cat /tmp/output.txt",
    store=[StoreSpec(src="/tmp/output.txt", name="shell-output")],
)
```

## Kubernetes-specific

On Kubernetes, file artifacts are stored as files-artifacts-expander init containers:

```bash
# See init containers for a service pod
kubectl describe pod <pod-name> -n kt-<enclave-name> | grep -A10 "Init Containers"

# Check if files-artifacts-expander succeeded
kubectl logs <pod-name> -n kt-<enclave-name> -c files-artifact-expander

# If the expander image is failing (ImagePullBackOff), check image tag
kubectl describe pod <pod-name> -n kt-<enclave-name> | grep "files-artifacts-expander"
```

## Common issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| File not found in service | Wrong mount path | Check `files` dict key matches expected path |
| Empty file after render | Template syntax error | Download artifact and inspect rendered output |
| Init container crash | files-artifacts-expander image issue | Check init container logs with kubectl |
| Artifact name conflict | Duplicate artifact names | Use unique names for each `plan.upload_files()` / `plan.render_templates()` |
| Permission denied | Container runs as non-root | Mount to a writable path or adjust image permissions |
