---
title: Roadmap
sidebar_label: Roadmap
slug: '/roadmap'
---

:::tip
Kurtosis is rapidly evolving alongside the needs of our users. As a result, please interpret the below to be accurate for approximately 3 months from the last updated date.

The last updated date is **August 1, 2023**
:::

Over the next 1-3 months, we will be making investments in our product to enable workflows that involve long-lived environments. Doing so cements the value proposition that we offer for both dev and test, and opens a new world when it comes to production use cases. Directionally, these efforts represent a step closer to our goal of extending Kurtosis across the entire development lifecycle. Our investments will be spread across various features and improvements but will generally fall into one of the below buckets:

- **More robust support for various workflows involving enclaves deployed on Kubernetes.** This includes support for graceful blue/green rollouts, support for replication controllers like ReplicaSet (RS), and cleaner ways to interact with the cluster from the outside.
- **Idempotent runs** that enable a developer to make changes to the Starlark package & Kurtosis will apply those changes to a long-lived enclave deterministically.
- **Frontend improvements** to support a cleaner interface for users to deploy & interact with a long-lived enclave in the cloud. These improvements would be in service of ensuring that the experience is as seamless and self-service as possible.
- **Connectivity to and from long-lived enclaves** to standardize, both from a user experience and technically, how one would get traffic to and from the enclave. This scope of work will include making it seamless to set up and manage the connection as well.
- **Persisting data** across enclaves, services within those enclaves, and beyond the lifecycle of a service and enclave as well. This includes the supporting workflows for trivial manipulation of the data inside a container.
- **Centralized logging infrastructure** to aggregate logs from everything inside an enclave, making them easily queryable, and storing them somewhere so that they can be used beyond the life of the enclave.
- **A fully managed cloud offering and accompanying self-service workflows** for a stress-free, easy way to deploy test and dev environments, that live as long as you need them to, directly onto remote infrastructure.

If any of the investments we are making interest you or if you have feedback for us, please let us know in our [Github Discussions](https://github.com/kurtosis-tech/kurtosis/discussions/categories/q-a) page, where we are fielding some great questions from our community.
