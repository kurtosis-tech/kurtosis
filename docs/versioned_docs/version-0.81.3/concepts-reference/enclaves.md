---
title: Enclaves
sidebar_label: Enclaves
---

An enclave is a Kurtosis primitive representing an isolated, ephemeral environment - like the Kurtosis version of a Kubernetes namespace. They are managed by the `enclave` family of CLI commands (e.g. [`kurtosis enclave add`][enclave-add-reference], [`kurtosis enclave ls`][enclave-ls-reference], [`kurtosis enclave inspect`][enclave-inspect-reference], etc.).

Each enclave houses an arbitrary number of services and [files artifacts][files-artifacts-reference]. The contents of an enclave are manipulated using [Starlark][starlark-reference] via [functions on the `Plan` object](../starlark-reference/plan.md)

When an enclave is removed via [`kurtosis enclave rm`][enclave-rm-reference] or [`kurtosis clean`][clean-reference], everything inside of it is destroyed as well.

<!----------------- ONLY LINKS BELOW HERE ------------------------------>
[enclave-add-reference]: ../cli-reference/enclave-add.md
[enclave-ls-reference]: ../cli-reference/enclave-ls.md
[enclave-inspect-reference]: ../cli-reference/enclave-inspect.md
[enclave-rm-reference]: ../cli-reference/enclave-rm.md
[clean-reference]: ../cli-reference/clean.md
[files-artifacts-reference]: ./files-artifacts.md
[plan-starlark-reference]: ../starlark-reference/plan.md
[starlark-reference]: ./starlark.md
