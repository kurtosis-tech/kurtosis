---
title: cluster get
sidebar_label: cluster get
slug: /cluster-get
---

Kurtosis will work locally or over remote infrastructure. To determine the type of infrastructure your instance of Kurtosis is running on, simply run:

```bash
kurtosis cluster get
```

The clusters that Kurtosis can connect to are defined in your `kurtosis-config.yml` file, located at `/Users/<YOUR_USER>/Library/Application Support/kurtosis/kurtosis-config.yml` on MacOS. See [this guide](https://docs.kurtosis.com/k8s#iii-add-your-cluster-information-to-kurtosis-configyml) to learn more about how to add cluster information to your `kurtosis-config.yml` file.

Below is an example of what a valid `kurtosis-config.yml` file might look like with the clusters: `docker`, `minikube`, and `cloud`:
```yml
config-version: 2
should-send-metrics: true
kurtosis-clusters:
  docker:
    type: "docker"
  minikube:
    type: "kubernetes"
    config:
      kubernetes-cluster-name: "minikube"
      storage-class: "standard"
      enclave-size-in-megabytes: 10
  cloud:
    type: "kubernetes"
    config:
      kubernetes-cluster-name: "NAME-OF-YOUR-CLUSTER"
      storage-class: "standard"
      enclave-size-in-megabytes: 10
```