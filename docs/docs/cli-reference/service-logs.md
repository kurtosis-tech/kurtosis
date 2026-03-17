---
title: service logs
sidebar_label: service logs
slug: /service-logs
---

To print the logs for services in an enclave, run:


```bash
kurtosis service logs $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER1 $THE_SERVICE_IDENTIFIER2 $THE_SERVICE_IDENTIFIER3
```
where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and services, respectively. The service identifier (name or UUID) is printed upon inspecting an enclave. 
:::

:::note Number of log lines
By default, logs printed in the terminal from this command are truncated at the most recent 200 log lines. For a stream of logs, we recommend the `-f` flag. For all the logs use the `-a` flag and for a snapshot of the logs at a given point in time (e.g. after a change), we recommend the [`kurtosis dump`](./dump.md).
:::

:::note Log Retention
Kurtosis will keep logs for up to 1 week before removing them to prevent logs from taking up to much storage. If you'd like to remove logs before the retention period, `kurtosis enclave rm` will remove any logs associated for service in the enclave and `kurtosis clean` will remove logs for all services in stopped enclaves.
:::

:::note Kubernetes Logging
Kurtosis logging on the Kubernetes backend requires using kurtosis config v4. Kurtosis config v4 introduces a required field in the Kubernetes cluster configuration: `engine-node-name`. This field instructs Kurtosis to schedule the engine and logs aggregator pods on the requested node. The logs aggregator will output logs to the filesystem of `engine-node-name`, where the engine reads them for `service logs`. Kurtosis will always attempt to schedule these pods on the `engine-node-name` and will fail if it is unable to do so.
```
config-version: 4 # set config version to v4
should-send-metrics: true
kurtosis-clusters:
    minikube:
        type: "kubernetes"
        config:
        kubernetes-cluster-name: "minikube"
        enclave-size-in-megabytes: 10
        storage-class: "standard"
        engine-node-name: "minikube-m02" # additional required config in v4
...
```
:::

The following optional arguments can be used:
1. `-a`, `--all` can be used to retrieve all logs.
1. `-n`, `--num=uint32` can be used to retrieve X last log lines. (eg. `-n 10` will retrieve last 10 log lines, similar to `tail -n 10`)
1. `-f`, `--follow` can be added to continue following the logs, similar to `tail -f`.
1. `-x`, `--all-services` can be used to retrieve logs for all services in an enclave. Another option is to pass in the escaped wildcard operator like so `kurtosis service logs enclave-name '*'`
1. `--match=text` can be used for filtering the log lines containing the text.
1. `--regex-match="regex"` can be used for filtering the log lines containing the regex. This filter will also work for text but will have degraded performance.
1. `-v`, `--invert-match` can be used to invert the filter condition specified by either `--match` or `--regex-match`. Log lines NOT containing the match will be returned.

Important: `--match` and `--regex-match` flags cannot be used at the same time. You should either use one or the other.
