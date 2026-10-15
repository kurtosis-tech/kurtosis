---
title: otel start
sidebar_label: otel start
slug: /otel-start
---

To start a local OpenTelemetry collector and ClickHouse instance in Docker, run:

```bash
kurtosis otel start
```

This command starts two engine-managed side containers — an [OpenTelemetry collector][otel-collector] and a single-node [ClickHouse][clickhouse] — and restarts the Kurtosis engine with an updated configuration. The new configuration adds a log sink that routes enclave logs to the collector over Loki's HTTP protocol; the collector then writes them to ClickHouse, where they can be queried alongside any OTLP traces and metrics you push to the collector from inside your enclaves.

The side containers publish the following ports on the Docker host so they are reachable from inside every enclave as well as from tools on your machine:

| Service               | Host port | Purpose                       |
|-----------------------|-----------|-------------------------------|
| OpenTelemetry — OTLP gRPC | `14317`   | OTLP gRPC ingest              |
| OpenTelemetry — OTLP HTTP | `14318`   | OTLP HTTP ingest              |
| ClickHouse — HTTP     | `18123`   | ClickHouse HTTP query/insert  |

The host ports are deliberately non-default (e.g. `14317` instead of `4317`) so the side containers do not collide with a developer's own ClickHouse or OTLP collector already bound on the Docker host.

Once the side containers are running, enclave logs are tagged with the enclave name (not just the enclave UUID), so logs and traces can be tenanted and queried by enclave name.

`kurtosis otel start` is **Docker-only** — it returns an error on Kubernetes and Podman backends. It is also mutually exclusive with [grafloki start][grafloki-start]: running `kurtosis otel start` skips any configured Grafana/Loki setup, and the engine will only export logs to the OpenTelemetry collector while the otel side containers are running.

On Docker, the OpenTelemetry collector is the default: `kurtosis engine start` and `kurtosis engine restart` auto-start the side containers and wire up the Loki sink, so `kurtosis otel start` is only needed after an `otel stop`. To opt out, set [`backend-log-collector: vector`][kurtosis-config] in `kurtosis-config.yml`:

```yaml
config-version: 9
should-send-metrics: true
kurtosis-clusters:
  docker:
    type: docker
    backend-log-collector: vector
```

The default also falls back to Vector when `grafana-loki.should-start-before-engine: true` is set.

To stop the side containers until the next engine start, use [otel stop][otel-stop].

Read more about sinks and how to [export logs][export-logs] from Kurtosis.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[otel-collector]: https://opentelemetry.io/docs/collector/
[clickhouse]: https://clickhouse.com/
[grafloki-start]: ./grafloki-start.md
[otel-stop]: ./otel-stop.md
[export-logs]: ../guides/exporting-logs.md
[kurtosis-config]: ../advanced-concepts/kurtosis-config.md
