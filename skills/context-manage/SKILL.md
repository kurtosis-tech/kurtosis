---
name: context-manage
description: Manage Kurtosis contexts for connecting to different Kurtosis instances. Add, list, switch, and remove contexts. Use when working with multiple Kurtosis environments (local, remote, team shared).
compatibility: Requires kurtosis CLI.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Context Manage

Manage Kurtosis contexts for connecting to different Kurtosis instances.

## What are contexts?

Contexts define which Kurtosis instance the CLI talks to. The default context is `default` which connects to the local engine (Docker or Kubernetes).

## List contexts

```bash
kurtosis context ls
```

Shows all configured contexts and which one is active.

## Add a context

```bash
kurtosis context add <context-name>
```

## Switch context

```bash
kurtosis context set <context-name>
```

After switching, restart the engine or portal as needed.

## Remove a context

```bash
kurtosis context rm <context-name>
```

## Common workflow

```bash
# Check which context is active
kurtosis context ls

# Switch to a different environment
kurtosis context set staging

# Start portal if using a remote context
kurtosis portal start

# Switch back to local
kurtosis context set default
```
