---
name: portal
description: Manage Kurtosis Portal for remote context access. Start, stop, and check status of the Portal daemon that enables communication with remote Kurtosis servers. Use when working with remote Kurtosis contexts.
compatibility: Requires kurtosis CLI.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Portal

Manage the Kurtosis Portal daemon for remote context access.

## What is Portal?

Kurtosis Portal is a lightweight local daemon that enables communication with Kurtosis enclaves running on a remote Kurtosis server. It's only needed when using remote contexts — not required for local Docker or direct Kubernetes access.

## Start

```bash
kurtosis portal start
```

## Check status

```bash
kurtosis portal status
```

## Stop

```bash
kurtosis portal stop
```

## When you need it

Portal is used with remote Kurtosis contexts. If you're using:
- **Local Docker**: No portal needed
- **Direct Kubernetes**: Use `kurtosis gateway` instead
- **Remote Kurtosis server**: Use portal + remote context

## Remote contexts

```bash
# List contexts
kurtosis context ls

# Add a remote context
kurtosis context add <context-name>

# Switch to remote context
kurtosis context set <context-name>

# Start portal for the remote context
kurtosis portal start
```
