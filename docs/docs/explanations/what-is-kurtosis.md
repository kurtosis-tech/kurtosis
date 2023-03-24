---
title: What Is Kurtosis and Why?
sidebar_label: What Is Kurtosis and Why?
sidebar_position: 1
---

### What is Kurtosis?
[Kurtosis](https://www.kurtosis.com) is a composable build system for reproducible test environments, enabling developers to easily define dynamic service dependencies, inject data, and re-use environment definitions for multi-container tests. 

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

### Why Kurtosis?
Our goal with Kurtosis is to make building a distributed application as easy as developing a single-server app. We aim to realize this goal by making it easier to configure multi-container test environments. 

Specifically, Kurtosis was built to tackle the following difficulties when it comes ot building distributed systems:
- Setting up test environments that have dynamic dependencies between services
- Reusing test environment definitions across different scenarios
- Injecting data into test environments for use across different types of tests

Our philosophy is that the distributed nature of modern software means that modern software development now happens at the environment level. Spinning up a single service container in isolation is difficult because it has implicit dependencies on other resources: services, volume data, secrets, certificates, and network rules. Therefore, the environment - not the container - is the fundamental unit of modern software.

This fact becomes apparent when we look at the software development lifecycle. Developers used to write code on their machine and ship a large binary to a few long-lived, difficult-to-maintain environments like Prod or Staging. Now, the decline of on-prem hardware, rise of containerization, and availability of flexible cloud compute enable the many environments of today: Prod, pre-Prod, Staging, Dev, and even ephemeral preview, CI test, and local dev.

The problem is that our tools are woefully outdated. The term "DevOps" was coined during the Agile revolution in the early 2000's. It signified making Dev responsible for end-to-end software delivery, rather than building software and throwing it over the wall to Ops to run. The idea was to shorten feedback loops, and it worked. However, our systems have become so complex that companies are now hiring "DevOps engineers" to manage the Docker, AWS, Terraform, and Kubernetes underlying all modern software. Though we call them "DevOps engineers", we are recreating Ops and separating Dev and Ops once more. 

In our vision, a developer should have a single platform for prototyping, testing, debugging, deploying to Prod, and observing the live system. Our goal with Kurtosis is to bring DevOps back.

To read more about our beliefs on reusable environments, [go here][reusable-environment-definitions]. To get started using Kurtosis, see [the quickstart][quickstart].

<!------------------------- REFERENCE LINKS ONLY ------------------------------------>
[reusable-environment-definitions]: ./reusable-environment-definitions.md
[quickstart]: ../quickstart.md
