---
title: Enclaves
sidebar_label: Enclaves
---

An enclave is a Kurtosis primitive representing an isolated, ephemeral environment - like the Kurtosis version of a Kubernetes namespace. They are managed by the `enclave` family of CLI commands (e.g. [`kurtosis enclave add`][enclave-add-reference], [`kurtosis enclave ls`][enclave-ls-reference], [`kurtosis enclave inspect`][enclave-inspect-reference], etc.).

Each enclave houses an arbitrary number of services and [files artifacts][files-artifacts-reference]. The contents of an enclave are manipulated via [Starlark instructions][starlark-instructions-reference] on [the `plan` object][plan-reference].

When an enclave is removed via [`kurtosis enclave rm`][enclave-rm-reference] or [`kurtosis clean`][clean-reference], everything inside of it is destroyed as well.

<!----------------- ONLY LINKS BELOW HERE ------------------------------>
[enclave-add-reference]: ./cli/enclave-add.md
[enclave-ls-reference]: ./cli/enclave-ls.md
[enclave-inspect-reference]: ./cli/enclave-inspect.md
[enclave-rm-reference]: ./cli/enclave-rm.md
[clean-reference]: ./cli/clean.md
[files-artifacts-reference]: ./files-artifacts.md
[starlark-instructions-reference]: ./starlark-instructions.md
[plan-reference]: ./plan.md
