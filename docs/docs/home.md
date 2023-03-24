---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---
## What is Kurtosis?
Kurtosis is a composable build system for multi-container test environments. 

#### Kurtosis has a definition language with:
- An instruction set of useful primitives for setting up and manipulating environment
- A scriptable Python-like SDK in Starlark, a build language used by Google’s Bazel
- A package management system for shareability and composability

#### Kurtosis has a validator for:
- Compile-time safety to quickly catch errors in test environment definitions
- The ability to dry-run test environment definitions to verify what will be run, before running

#### Kurtosis has a runtime to:
- Run multi-container test environments over Docker or Kubernetes, depending on how you wish to scale
- Enable debugging and investigation of problems live, as they're happening in your test environment, with an introspective toolkit
- Manage file dependencies to ensure tests environments are completely reproducible across different test runs and backends

## What problems does Kurtosis help solve?
Our goal with Kurtosis is to make building a distributed application as easy as developing a single-server app. We aim to realize this goal by making it easier to configure multi-container test environments. 

Specifically, Kurtosis was built to tackle the following difficulties when it comes ot building distributed systems:
- Setting up test environments that have dynamic dependencies between services
- Reusing test environment definitions across different scenarios
- Injecting data into test environments for use across different types of tests

#### Try out Kurtosis now

Try Kurtosis now with our [quickstart](./quickstart.md).

:::info
If you have questions, need help, or simply want to learn more, schedule some time with our [cofounder, Kevin](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).
:::


