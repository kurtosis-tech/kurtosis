---
title: otel stop
sidebar_label: otel stop
slug: /otel-stop
---

To stop the OpenTelemetry side containers that were started by [otel start][otel-start], run:

```bash
kurtosis otel stop
```

This restarts a running Kurtosis engine without the OpenTelemetry Loki sink — reverting to the default log sinks — and then stops the OpenTelemetry collector and ClickHouse side containers, if they exist. If the engine is not running, the engine restart step is skipped and only the side containers are stopped.

Like [otel start][otel-start], this command is **Docker-only**.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[otel-start]: ./otel-start.md
