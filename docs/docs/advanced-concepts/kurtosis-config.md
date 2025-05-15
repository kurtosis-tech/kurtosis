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
# Latest supported version is 6, supported by Kurtosis engine version 1.9.0.
config-version: 6

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

    # Optional. Enables sending logs to a locally managed Grafana + Loki instance via `kurtosis grafloki start`.
    grafana-loki:
      grafana-image: "grafana/grafana:11.6.0"
      loki-image: "grafana/loki:2.9.4"

      # Starts Grafana and Loki before engine - useful if Grafana Loki is default logging setup
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