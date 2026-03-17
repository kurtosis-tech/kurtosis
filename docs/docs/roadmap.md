---
title: Roadmap
sidebar_label: Roadmap
slug: '/roadmap'
---

:::note
Last updated: May 5, 2025
:::

:::info
Kurtosis Technologies open sourced Kurtosis in June '24. Since then, Kurtosis has grown via open source contributions and active maintenance. If you have ideas to improve Kurtosis, please make a PR to suggest them under Kurtosis Improvement Proposals and get in touch with one of the [maintainers](https://github.com/kurtosis-tech/kurtosis/blob/main/MAINTAINERS.md).
:::

## Roadmap
Over the next 3–6 months, Kurtosis maintainers aim to improve the following product areas:

### **Persisting enclave data**

Enclaves contain valuable state that can aid in debugging, reproducing environments, and saving time when spinning up new setups. Today, Kurtosis captures much of an enclave's state (e.g., service info, file artifacts, persistent directories, running containers), but there’s no easy way to extract and reuse that data to reproduce an environment.

We’ll be exploring features like restarting enclaves, snapshotting enclave state, and injecting data into enclaves.

If this is relevant to you or your team's workflows, please reach out to [Tedi Mitiku](https://tedi.dev).

### **Faster local development loop**

As the ecosystem of Kurtosis packages grows, users are composing multiple packages into larger setups. Fortunately, Kurtosis makes composition easy—but the result is that these packages are getting bigger and taking longer to spin up locally for testing. Once the enclave is running, developers want to iterate quickly: modify service code and immediately see the changes reflected.

Potential improvements include enabling a watch mode that detects updated Docker images and reloads them into the enclave automatically, or supporting hot reload by detecting binary changes and applying them live.

### **Support for long-lived Kubernetes environments**

Kurtosis simplifies orchestration on Kubernetes, but it’s not yet optimized for managing environments that run for days or months. Enhancing support for Kubernetes-native features like StatefulSets and ReplicaSets is a priority.

Teams like [Bloctopus.io](https://www.bloctopus.io/) are actively contributing to this area to make Kurtosis better suited for persistent, long-lived use cases.

If this interests you, or if you have feedback, please reach out to the Bloctopus team on [TG](https://t.me/wanderosity) or anisha@bloctopus.io

### **Speeding up Kurtosis**

Kurtosis relies heavily on containers, launching them for nearly every task. While this design provides flexibility, it adds overhead—especially for lightweight tasks like parsing or ETL, where container startup and teardown create delays.

Possible optimizations include reusing a dedicated container for lightweight tasks or supporting task execution outside of containers entirely.

If any of these investments interest you, or if you have feedback, please let us know in our [GitHub Discussions](https://github.com/kurtosis-tech/kurtosis/discussions/categories/q-a), [Discord](https://discord.com/invite/TMhR2uX5WMZ), or feel free to reach out directly to a maintainer or [Tedi Mitiku](https://tedi.dev).

## **Kurtosis Improvement Proposals**

We encourage users to fork Kurtosis and improve the engine in ways that suit their needs. For larger features that require discussion and coordination, we welcome proposals via Kurtosis Improvement Proposals (KIPs).

## **KIP-001 - Persistent, Scalable Devnets**
Kurtosis began as a developer‑focused tool to spin up reproducible, ephemeral environments. Over the last two years the community has stretched those limits, running long‑lived devnets and complex integration testbeds on top of Kurtosis. 
We propose an evolution that promotes Kurtosis from purely ephemeral to persistent devnets, and eventually, a production‑grade orchestration platform capable of hosting stateful, durable, multi‑tenant networks in the cloud.

[KIP-001 for Persistent, Scalable Devnets detailed here](https://daffodil-porpoise-366.notion.site/KIP-001-Kurtosis-Improvement-Proposal-1df1019e456a8056b051f908775da808)

If this interests you, or if you have feedback, please reach out to the Bloctopus team on [TG](https://t.me/wanderosity) or anisha@bloctopus.io

If you'd like to propose a feature, please create a document and submit a pull request to add it to the KIP list!
