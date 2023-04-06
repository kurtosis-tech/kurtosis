---
title: Starlark
sidebar_label: Starlark
---

[Starlark](https://github.com/bazelbuild/starlark) is a minimal programming language, halfway between a configuration language and a general-purpose programming language. It was developed by Google to do configurations for the [Bazel build tool](https://bazel.build/rules/language), and has since [been adopted by Facebook for the Buck build system as well](https://github.com/facebookexperimental/starlark-rust). Starlark's syntax is a minimal subset of of Python, with a focus on readability. [This page](https://bazel.build/rules/language#differences_with_python) lists the differences between Starlark and Python.

Kurtosis uses Starlark as the way for users to express manipulations to [enclave][enclaves-reference]. Users submit Starlark scripts to Kurtosis, the Starlark is interpreted, and the instructions in the script are executed.

To read more about why we chose Starlark, see [this page][why-kurtosis-starlark].

How is Starlark implemented at Kurtosis?
----------------------------------------
Starlark itself is very basic; Google designed it to be extended to fulfill a given usecase. For example, the Bazel language is actually an extension built on top of Starlark. 

We extended basic Starlark with [our own DSL](../starlark-reference/index.md) so that it could [fulfill the properties of reusable environment definitions](../explanations/reusable-environment-definitions.md). This gave us:

- A [list of Kurtosis-specific functions][starlark-reference] for working with an environment
- The [ability to accept parameters][run-args-reference]
- Dependencies, so Kurtosis scripts can [import other scripts][locators-reference]
- A [GitHub-based packaging system](./packages.md), so environment definitions can be shared with each other

Additionally, we built a [multi-phase engine][multi-phase-runs-reference] around the Starlark interpreter, to provide [users with benefits not normally available in a scripting language][multi-phase-runs-explanation].

:::tip Visual Studio Code (VS Code) Extension
We've released an [official Kurtosis Starlark VS Code extension][vscode-plugin] to enrich the developer experience when writing packages with Starlark. Features include: syntax highlighting, method signature suggestions, hover preview for functions, and auto-completion for Kurtosis custom types.
:::

<!--------------- ONLY LINKS BELOW HERE --------------------------->
[enclaves-reference]: ./enclaves.md
[why-kurtosis-starlark]: ../explanations/why-kurtosis-starlark.md
[starlark-reference]: ../starlark-reference/index.md
[run-args-reference]: ./packages.md#arguments
[locators-reference]: ./locators.md
[multi-phase-runs-reference]: ../concepts-reference/multi-phase-runs.md
[multi-phase-runs-explanation]: ../explanations/why-multi-phase-runs.md
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
