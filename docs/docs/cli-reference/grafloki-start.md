---
title: grafloki start
sidebar_label: grafloki start
slug: /grafloki-start
---

To start a Grafana/Loki instance in Docker or K8s cluster, run:

```bash
kurtosis grafloki start
```

This command starts a local Grafana/Loki instance and restarts the Kurtosis engine with an updated configuration. The new configuration includes a log sink that routes logs to the local Grafana/Loki instance.

Read more about sinks and how to [export logs][export-logs] from Kurtosis.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[export-logs]: ../guides/exporting-logs.md