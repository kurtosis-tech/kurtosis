---
title: Kurtosis Config
sidebar_label: Kurtosis Config
---

# Kurtosis Config

This file defines how Kurtosis behaves across clusters and environments. You can use it to configure logging, telemetry, and advanced cluster-specific settings.

Below is a fully annotated example of a `kurtosis-config.yml` with explanations for each field.

```yaml
# kurtosis-config.yml

# Required. The version of the Kurtosis config schema.
# This ensures compatibility with the CLI. 
# Latest supported version is 9.
config-version: 9

# Optional. Whether Kurtosis should send anonymous telemetry (usage) data.
# Default: true
# Set to false to opt out.
should-send-metrics: true

# Optional. Defines configurations for one or more Kurtosis clusters.
# Each key is a user-defined cluster name.
kurtosis-clusters:
  docker:  # Name of the cluster, this can be anything and is used to identify clusters in `kurtosis cluster set/get`
    # Required. Determines the cluster type.
    # Valid values: "docker", "kubernetes", "podman"
    type: docker

    # Optional. Controls whether the built-in logs DB (PersistentVolumeLogsDB) is enabled.
    # Default: true
    # Set to false if you're using an external logging system like Loki or Elasticsearch to save storage.
    should-enable-default-logs-sink: true

    # Optional. Allows Docker-only ServiceConfig.privileged, ServiceConfig.bind_mounts,
    # and ServiceConfig.host_pid_namespace fields for CLI runs against this cluster. Default: false.
    # This is a CLI/request opt-in, not an engine-side operator policy. Direct API
    # clients can also opt in by setting allow_privileged_mode on the run request.
    allow-privileged-mode: false

    # Optional. Selects which log-collector stack the engine wires up at start.
    # Valid values: "vector" (default), "otel".
    # When set to "otel" (Docker only), `kurtosis engine start`/`restart` auto-starts the
    # OpenTelemetry collector and ClickHouse side containers and configures the engine's
    # Vector aggregator to ship logs to the collector via Loki HTTP. Equivalent to running
    # `kurtosis otel start` before `kurtosis engine start`.
    # Mutually exclusive with `grafana-loki.should-start-before-engine: true` below.
    backend-log-collector: vector

    # Optional. Configures external sinks to export service logs from enclaves.
    # This uses Vector under the hood and supports all Vector sink types.
    logs-aggregator:
      sinks:
        elasticsearch:
          type: elasticsearch
          bulk:
            index: "kt-{{ kurtosis_enclave_uuid }}-{{ kurtosis_service_uuid }}"
          auth:
            strategy: basic
            user: elastic
            password: "<PASSWORD>"
          tls:
            verify_certificate: false
          endpoints:
            - "https://<ELASTICSEARCH_IP_ADDRESS>:9200"

    # Optional. Enables advanced log filtering or transformation before logs are sent to sinks.
    # Uses Fluent Bit-style filters.
    logs-collector:
      filters:
        - name: grep
          match: "*"
          params:
            - key: exclude
              value: "log .*DEBUG.*"
        - name: modify
          match: "*"
          params:
            - key: Add
              value: "timestamp ${time}"

    # Optional. Configures the locally managed Loki / Grafana+Loki helpers used by
    # `kurtosis loki start` and `kurtosis grafloki start`.
    grafana-loki:
      grafana-image: "grafana/grafana:11.6.0"
      loki-image: "grafana/loki:2.9.4"

      # Starts Grafana and Loki before engine - useful if Grafana + Loki is the default logging setup
      should-start-before-engine: true 

  kube:  # A named Kubernetes cluster
    type: kubernetes

    # Optional. Kubernetes-specific configuration.
    config:
      kubernetes-cluster-name: "minikube"
      storage-class: "standard"
      enclave-size-in-megabytes: 10 

      # Name of node to schedule the engine and logs aggregator on.
      # Currently, the engine and logs aggregator will be scheduled on the same machine as they need to share a filesystem for reading and writing to default logs db.
      engine-node-name: "minikube-one"

      # Optional. Node selectors to apply to Kurtosis-managed pods (engine, logs aggregator).
      # These are merged with the engine-node-name selector if both are specified.
      node-selectors:
        disktype: ssd

      # Optional. Tolerations to apply to Kurtosis-managed pods (engine, logs aggregator, logs collector).
      # Allows scheduling on nodes with matching taints.
      tolerations:
        - key: "dedicated"
          operator: "Equal"
          value: "kurtosis"
          effect: "NoSchedule"

# Optional. Used when connecting to Kurtosis Cloud.
# Typically only needed in enterprise or managed deployments.
cloud-config:
  api-url: "https://api.kurtosis.cloud"
  port: 443
  certificate-chain: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
```

## Notes

- Kurtosis merges your config with internal defaults, so you only need to specify overrides.
- To see where your current config file is located, run:
  ```bash
    kurtosis config path  
  ```
