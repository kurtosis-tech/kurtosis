---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---

[Kurtosis](https://www.kurtosis.com) is a development platform for distributed applications that aims to provide a consistent experience across all stages of distributed app software delivery.

Use cases for Kurtosis include:

- Running a third-party distributed app, without knowing how to set it up
- Local prototyping & development on distributed apps
- Writing integration and end-to-end distributed app tests (e.g. happy path & sad path tests, load tests, performance tests, etc.)
- Running integration/E2E distributed app tests
- Debugging distributed apps during development

## Why Kurtosis?

Docker and Kubernetes are each great at serving developers in different parts of the development cycle: Docker for development/testing, Kubernetes for production. However, the separation between the two entails different distributed app definitions, and different tooling. In dev/test, this means Docker Compose and Docker observability tooling. In production, this means Helm definitions and manually-configured observability tools like Istio, Datadog, or Honeycomb.

![Why Kurtosis](@site/static/img/home/kurtosis-utility.png)

Kurtosis aims at one level of abstraction higher. Developers can define their distributed applications in Kurtosis, and Kurtosis will handle:

- Running on Docker or Kubernetes
- Reproduceability
- Safety
- Port-forwarding & local development hookups
- Observability
- Sharing

If we succeed in our vision, you will be able to use the same distributed application definition from local dev all the way to prod.
