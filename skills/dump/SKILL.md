---
name: dump
description: Dump Kurtosis state for debugging and sharing. Export enclave state including service logs, configurations, and file artifacts to a local directory. Use when you need to capture state for offline analysis or to share with others for debugging.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Dump

Export Kurtosis state for debugging and sharing.

## Dump entire Kurtosis state

```bash
kurtosis dump /tmp/kurtosis-dump
```

This exports everything: engine state, all enclaves, services, logs.

## Dump a specific enclave

```bash
kurtosis enclave dump <enclave-name> /tmp/enclave-dump
```

## What gets exported

The dump directory contains:
- **Service logs** — stdout/stderr from each service
- **Service configs** — how each service was configured
- **File artifacts** — all files stored in the enclave
- **Enclave metadata** — creation time, status, parameters

## Directory structure

```
/tmp/enclave-dump/
  service-1/
    spec.json         # Service configuration
    output.log        # Service logs
  service-2/
    spec.json
    output.log
  files-artifacts/
    artifact-name/
      file1.yaml
      file2.json
```

## Common uses

### Share a bug report

```bash
# Dump the problematic enclave
kurtosis enclave dump failing-enclave /tmp/bug-report

# Compress for sharing
tar czf bug-report.tar.gz -C /tmp bug-report
```

### Compare two runs

```bash
kurtosis enclave dump run-1 /tmp/dump-1
kurtosis enclave dump run-2 /tmp/dump-2
diff -r /tmp/dump-1 /tmp/dump-2
```

### Capture state before cleanup

```bash
# Dump everything before cleaning
kurtosis dump /tmp/pre-clean-dump
kurtosis clean -a
```
