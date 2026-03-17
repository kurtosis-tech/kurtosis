---
title: Using Claude Code Skills
sidebar_label: Claude Code Skills
slug: /claude-code-skills
sidebar_position: 16
---

This repository ships with a set of [Claude Code](https://claude.ai/download) skills — structured prompts that teach Claude how to build, debug, deploy, and manage Kurtosis. When loaded, Claude gains operational knowledge about the Kurtosis CLI, Docker and Kubernetes backends, Starlark development, and more.

## What are skills?

Skills are Markdown files (`SKILL.md`) in the `skills/` directory of the Kurtosis repository. Each skill contains step-by-step instructions, common flags, troubleshooting tips, and best practices for a specific area of Kurtosis. Claude Code automatically discovers and indexes these files so it can reference them during conversations.

## Available skills

### Core operations

| Skill | Description |
|-------|-------------|
| `clean` | Clean up enclaves and artifacts |
| `engine-manage` | Start, stop, restart the engine and check health |
| `cluster-manage` | Switch between Docker and Kubernetes backends |
| `context-manage` | Manage contexts for multiple Kurtosis environments |
| `run-package` | Run Starlark scripts and packages with all flags |

### Enclave and service management

| Skill | Description |
|-------|-------------|
| `enclave-inspect` | List enclaves, view services, ports, and file artifacts |
| `service-manage` | Add, stop, start, remove services; view logs and shell in |
| `files-inspect` | Inspect, download, upload, and debug file artifacts |
| `port-forward` | View and manage port mappings for services |
| `dump` | Export enclave state for offline debugging |

### Development and building

| Skill | Description |
|-------|-------------|
| `starlark-dev` | Write and debug Starlark packages from scratch |
| `cli-local-build` | Build and test the CLI from source |
| `docker-local-build` | Build all components and Docker images locally |
| `lint` | Lint and format Starlark files |
| `import-compose` | Convert Docker Compose files to Starlark packages |

### Kubernetes

| Skill | Description |
|-------|-------------|
| `k8s-dev-deploy` | Build, push, and deploy dev images to a K8s cluster |
| `k8s-debug-pods` | Diagnose Pending, CrashLoopBackOff, and scheduling issues |
| `k8s-clean-cluster` | Force-clean orphaned Kurtosis resources from a cluster |
| `gateway` | Start the gateway for forwarding ports to K8s services |

### Debugging and observability

| Skill | Description |
|-------|-------------|
| `docker-debug` | Inspect engine, APIC, and service logs on Docker |
| `grafloki` | Start Grafana and Loki for centralized log collection |
| `portal` | Manage the Portal daemon for remote context access |

## Installing skills

Skills are auto-discovered by Claude Code when present in a project's `skills/` directory. There are several ways to set them up depending on your workflow.

### Already cloning the repo (contributors)

If you're working in the Kurtosis repository, skills are already available — no extra setup needed. Claude Code discovers them automatically when you open a conversation in the repo directory.

### Copy skills into another project

To give Claude Kurtosis knowledge inside a different project:

```bash
cp -r /path/to/kurtosis/skills /path/to/your-project/skills
```

### Symlink from another project

Keep skills in sync with the Kurtosis repo without duplicating files:

```bash
# Symlink the entire directory
ln -s /path/to/kurtosis/skills /path/to/your-project/skills

# Or symlink only the skills you need
mkdir -p /path/to/your-project/skills
ln -s /path/to/kurtosis/skills/run-package /path/to/your-project/skills/run-package
ln -s /path/to/kurtosis/skills/starlark-dev /path/to/your-project/skills/starlark-dev
```

### Install globally

Make Kurtosis skills available in all Claude Code sessions regardless of which project you're in:

```bash
cp -r /path/to/kurtosis/skills ~/.claude/skills
```

## Using skills

Once installed, invoke a skill as a slash command in Claude Code:

```
/clean              # Clean up enclaves
/run-package        # Run a Starlark package
/docker-debug       # Debug Docker containers
/k8s-debug-pods     # Debug Kubernetes pods
/starlark-dev       # Help writing Starlark packages
```

You can also reference skills naturally in conversation. For example, asking "help me debug why my pod is stuck in Pending" will cause Claude to pull in the relevant `k8s-debug-pods` skill context automatically.

### Combining skills

Skills compose well together. For example, a typical development workflow might use:

1. `/starlark-dev` — write a new Starlark package
2. `/run-package` — run and test the package
3. `/enclave-inspect` — verify the enclave looks correct
4. `/docker-debug` or `/k8s-debug-pods` — troubleshoot any issues
5. `/clean` — tear everything down when done

## Writing new skills

To add a new skill, create a `SKILL.md` file in a new subdirectory under `skills/`:

```
skills/
  my-new-skill/
    SKILL.md
```

The file should include YAML frontmatter with at minimum a `name` and `description`:

```yaml
---
name: my-new-skill
description: Short description of what this skill helps with and when to use it.
compatibility: Any prerequisites (e.g., "Requires kurtosis CLI with a running engine.")
metadata:
  author: your-name
  version: "1.0"
---

# My New Skill

Instructions, commands, and tips go here...
```

Keep skills focused on a single area. Prefer concrete commands and examples over abstract explanations.
