---
title: Why Kurtosis Starlark?
sidebar_label: Why Kurtosis Starlark?
---

Background
----------
Distributed systems are very complex, and made up of many components. This means there are many, many instantiation, configuration, and manipulation actions possible with a distributed system.

Kurtosis [aims to provide a single solution across Dev, Test, and Prod][what-we-built-kurtosis]. We therefore needed a consistent way to represent all the various ways to manipulate a distributed system across Dev, Test, and Prod.

Additionally, we believe that [anything intended to work across Dev, Test, and Prod must have certain properties][reusable-environment-definitions]. We therefore needed something that would fulfill these properties.

With these constraints in mind, we went searching for a way to represent distributed system manipulations.

Attempt 1: Declarative Languages
--------------------------------
We first looked at declarative languages like YAML, Jsonnet, Dhall, and CUE. 

To use them, we'd need to write our own DSL (and accompanying parser) on top of the language to do what we needed. We knew that our "parameterizability" requirement meant users would need conditional/looping logic, but we were unhappy with how difficult parameters, conditionals, and loops were in these languages. 

The conditionals and parameters in the CircleCI YAML DSL seem to be a cautionary tale of starting with a declarative language and adding logic constructs later. [Others](https://github.com/tektoncd/experimental/issues/185#issuecomment-535338943) seemed [to agree](https://solutionspace.blog/2021/12/04/every-simple-language-will-eventually-end-up-turing-complete/): when dealing with configuration, start with a Turing-complete language because you will eventually need it.

Attempt 2: General-Purpose Languages
------------------------------------
We next looked at letting users declare environment definitions in their preferred general-purpose language, like Pulumi. This would require a large effort from our side to support many different languages, but we would do it if it was the right choice. 

However, we ultimately rejected this option. We found that general-purpose programming languages:

1. **Are TOO powerful:** the user could inadvertently write code that was nondeterministic and hard to debug. 
1. **Are a security risk:** we'd need to run user code to construct an environment, and running arbitrary user code is a security risk in one general-purpose language let alone in various.
1. **Aren't friendly for the author:** to make an environment definition usable by any consumer, the author would have to bundle their definition inside a container. Containerization makes development more painful - the user must know about Dockerfiles, their best practices, and how to build them locally - and requires a CI job to publish the container images up to Dockerhub.
1. **Aren't friendly for the environment definition consumer:** a developer investigating a third-party environment definition could easily be faced with a language they're not familiar with. Worse, general-purpose languages have many patterns for accomplishing the same task, so the consumer would need to understand the class/object/function architecture.

Attempt 3: Starlark
-------------------
When we discovered Starlark, the fit was obvious. Starlark:

- Is syntactically valid Python, which most developers are familiar with (and Python syntax highlighting Just Works)
- [Intentionally removes many Python features][starlark-differences-with-python], to make Starlark easier to read and understand
- Has [several properties that are very useful for Kurtosis](https://github.com/bazelbuild/starlark#design-principles), thanks to Starlark's origin as a build system configuration language
- [Has been around in Google since at least 2017](https://blog.bazel.build/2017/03/21/design-of-skylark.html), meaning it's well-vetted
- Is used for both Google and Facebook's build system, meaning it isn't going away any time soon
- [Is used by several other companies beyond Google and Facebook](https://github.com/bazelbuild/starlark/blob/master/users.md#users)


Conclusion
----------
So far, both our users and our team have been very happy with our decision to go with Starlark. If you've never used Starlark before, [the quickstart][quickstart] will be a good introduction.

<!--------------- ONLY LINKS BELOW HERE --------------------------->
[what-we-built-kurtosis]: ./why-we-built-kurtosis.md
[reusable-environment-definitions]: ./reusable-environment-definitions.md
[starlark-differences-with-python]: https://bazel.build/rules/language#differences_with_python

[quickstart]: ../quickstart.md
