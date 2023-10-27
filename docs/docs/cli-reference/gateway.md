---
title: gateway
sidebar_label: gateway
slug: /gateway
---

The Kurtosis gateway command is exclusively for use with Kubernetes backends and is intended to be a command you invoke before interacting with a remote Kurtosis engine deployed on a Kubernetes clsuter. Assuming you have configured your `kurtosis-config.yml` file to work with your Kubernetes cluster [guide here](../guides/running-in-k8s.md), then this command will start a local "gateway" to connect your local machine to your remote Kubernetes cluster, enabling you to interact with the remote Kurtosis engine that lives on that cluster. 

```console
kurtosis gateway
```