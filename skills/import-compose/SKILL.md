---
name: import-compose
description: Import Docker Compose files into Kurtosis. Convert docker-compose.yml to Starlark packages or run them directly. Use when migrating existing Docker Compose workflows to Kurtosis.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Import Compose

Import Docker Compose files into Kurtosis, either running them directly or converting to Starlark.

## Run a Docker Compose file

```bash
kurtosis import docker-compose.yml
```

This creates an enclave and runs all services defined in the Compose file.

## Run with a named enclave

```bash
kurtosis import -n my-enclave docker-compose.yml
```

## Convert only (don't run)

Generate Starlark code from a Compose file without executing:

```bash
kurtosis import -c docker-compose.yml
```

This outputs Starlark that you can save and customize.

## Custom .env file

```bash
kurtosis import -e ./custom.env docker-compose.yml
```

Default is `.env` in the current directory.

## Supported Compose features

Most common Compose features are supported:
- Services with images
- Port mappings
- Environment variables
- Volumes (converted to file artifacts)
- Depends-on (converted to service ordering)
- Networks (Kurtosis handles networking automatically)

## Limitations

- Build directives are not supported (images must be pre-built)
- Some advanced networking features may not translate directly
- Volume mounts become file artifacts (not persistent volumes)
