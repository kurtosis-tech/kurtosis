---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---

## What is Kurtosis?
[Kurtosis](https://www.kurtosis.com) is a platform for packaging and launching environments of containerized services ("distributed applications") with a focus on approachability for the average developer. What Docker did for shipping binaries, Kurtosis aims to do even better for distributed applications. 

Kurtosis is formed of:

- A language for declaring a distributed application in Python syntax ([Starlark](https://github.com/google/starlark-go/blob/master/doc/spec.md))
- A packaging system for sharing and reusing distributed application components
- A runtime that makes a Kurtosis app Just Work, independent of whether it's running on Docker or Kubernetes, local or in the cloud
- A set of tools to ease common distributed app development needs (e.g. a log aggregator to ease log-diving, automatic port-forwarding to ease connectivity, a `kurtosis service shell` command to ease container filesystem exploration, etc.)

Go [here](../explanations/why-we-built-kurtosis.md) to learn more about what inspired us to build Kurtosis.

## Why should I use Kurtosis?
Kurtosis shines when creating, working with, and destroying self-contained distributed application environments. Currently, our users report this to be most useful when:

- You're developing on your application and you need to rapidly iterate on it
- You want to try someone's containerized service or distributed application without setting up an environment, dependencies, etc.
- You want to spin up your distributed application in ephemeral environments as part of your integration tests
- You want to ad-hoc test your application on a big cloud cluster
- You're the author of a containerized service or distributed application and you want to give your users a one-liner to try it
- You want to get an instance of your application running in the cloud without provisioning or administering a Kubernetes cluster

If you're in web3, we have even more specific web3 usecases [here](https://web3.kurtosis.com).

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

## Try out Kurtosis now

Try Kurtosis now with our [quickstart](./quickstart.md).

:::info
If you have questions, need help, or simply want to learn more, schedule a live session with us, go [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).
:::
