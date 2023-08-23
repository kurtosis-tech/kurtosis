---
title: Running Kurtosis in Kurtosis Cloud
sidebar_label: Running in Kurtosis Cloud
slug: /cloud
---

Kurtosis Cloud is a fully managed cloud offering and accompanying self-service workflows for a stress-free, easy way to deploy test and dev environments, that live as long as you need them to, directly onto remote infrastructure. By logging into our [cloud portal](https://cloud.kurtosis.com), a cloud instance will be provisioned to run your test and dev enclaves.

You can interact with your enclaves using the UI (or the [CLI](./installing-the-cli.md#ii-install-the-cli) for more advanced use cases).

![enclave-manager-ui](/img/guides/enclave-manager-ui.png)

A Kurtosis cloud instance is an AWS EC2 instance running the Kurtosis engine, the Kurtosis API controller and your enclave services inside Docker.  The service ports are forwarded to your local machine.

![cloud-arch](/img/guides/cloud-arch.png)
