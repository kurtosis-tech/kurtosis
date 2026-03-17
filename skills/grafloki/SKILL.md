---
name: grafloki
description: Start Grafana and Loki for centralized log collection from Kurtosis enclaves. View aggregated service logs in a Grafana dashboard. Use when you need a UI for browsing logs across multiple services or want persistent log storage.
compatibility: Requires kurtosis CLI with a running engine.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Grafloki

Start a Grafana + Loki stack for centralized log collection from Kurtosis enclaves.

## Start

```bash
kurtosis grafloki start
```

This creates a Grafana instance with Loki as a data source, collecting logs from all services in all enclaves.

## Stop

```bash
kurtosis grafloki stop
```

## Usage

After starting, open the Grafana URL shown in the output. Use the Explore view with the Loki data source to query logs:

- Filter by service name
- Search for specific log patterns
- View logs across multiple services side by side

## When to use

- Debugging multi-service issues where you need correlated logs
- Monitoring long-running enclaves
- When `kurtosis service logs` isn't enough (need search, filtering, time ranges)
