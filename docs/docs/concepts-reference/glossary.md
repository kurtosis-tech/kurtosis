---
title: Glossary
sidebar_label: Glossary
---

<!-- NOTE TO KURTOSIS DEVS: KEEP THIS ALPHABETICALLY SORTED -->

### API Container
The container that runs inside of each enclave. Receives Starlark via API and manipulates the enclave according to the instructions in Starlark.

### CLI
A command line interface, [installed by your favorite package manager](../guides/installing-the-cli.md), which wraps the Kurosis Go [client library][client-libs-reference] to allow you to manipulate the contents of Kurtosis.

### Enclave
An environment, isolated from other enclaves, in which distributed systems are launched and manipulated.

### Engine
The Kurtosis engine which receives instructions via API (e.g. "launch this service in this enclave", "create a new enclave", "destroy this enclave", etc.).

### Locator
A URL-like string for referencing resources. Also see [the extended documentation][locators].

### Package
A directory containing [a `kurtosis.yml` file][kurtosis-yml] and any additional modules and static files that the package needs. Also see [the extended documentation][packages].

### Starlark
[A minimal, Python-like language invented at Google](https://github.com/bazelbuild/starlark) for configuring their build tool, Bazel.

### User Service
A container, launched inside an enclave upon a request to the Kurtosis engine, that is started from whatever image the user pleases.

<!-- NOTE TO KURTOSIS DEVS: KEEP THIS ALPHABETICALLY SORTED -->

<!--------------------- ONLY LINKS BELOW HERE --------------------------->
[locators]: ./locators.md
[kurtosis-yml]: ./kurtosis-yml.md
[packages]: ./packages.md
[client-libs-reference]: ../client-libs-reference.md
