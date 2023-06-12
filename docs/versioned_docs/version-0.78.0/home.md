---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---
## What is Kurtosis?
[Kurtosis](https://www.kurtosis.com) is a composable build system for multi-container test environments. Kurtosis makes it easier for developers to set up test environments that require dynamic setup logic (e.g. passing IPs or runtime-generated data between services) or programmatic data seeding.

Go [here](./explanations/why-we-built-kurtosis.md) to learn more what inspired us to build Kurtosis.

## Why use Kurtosis?

Developers usually set up these types of dynamic environments with a free-form scripting language like bash or Python, interacting with the Docker CLI or Docker Compose. Kurtosis is designed to make these setups easier to maintain and reuse in different test scenarios.

In Kurtosis, test environments have these properties:
- Environment-level portability: the entire test environment always runs the same way, regardless of the host machine
- Composability: environments can be composed and connected together without needing to know the inner details of each setup
- Parameterizability: environments can be parameterized, so that they're easy to modify for use across different test scenarios

## Architecture

#### Kurtosis has a definition language of:
- An instruction set of useful primitives for setting up and manipulating environments
- A scriptable Python-like SDK in Starlark, a build language used by Google’s Bazel
- A package management system for shareability and composability

#### Kurtosis has a validator with:
- Compile-time safety to quickly catch errors in test environment definitions
- The ability to dry-run test environment definitions to verify what will be run, before running

#### Kurtosis has a runtime to:
- Run multi-container test environments over Docker or Kubernetes, depending on how you wish to scale
- Enable debugging and investigation of problems live, as they're happening in your test environment
- Manage file dependencies to ensure complete portability of test environments across different test runs and backends

#### Try out Kurtosis now

Try Kurtosis now with our [quickstart](./quickstart.md).

:::info
If you have questions, need help, or simply want to learn more, schedule a live session with us, go [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).
:::
