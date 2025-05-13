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

Kurtosis Config allows configuring `grafana-loki` and `should-enable-default-logs-sink` configurations.

```yaml
config-version: 5
should-send-metrics: true
kurtosis-clusters:
  docker:
    type: "docker"
    grafana-loki:
      grafana-image: "grafana/grafana:11.6.0"
      loki-image: "grafana/loki:2.9.4"
      # Starts Grafana and Loki before engine - useful if Grafana Loki is default logging setup
      should-start-before-engine: true 
    # If set to false, Kurtosis will not collect logs in the default PersistentVolumeLogsDB
    # Useful to save storage if leveraging external logging setups
    should-enable-default-logs-sink: false
```

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[export-logs]: ../guides/exporting-logs.md