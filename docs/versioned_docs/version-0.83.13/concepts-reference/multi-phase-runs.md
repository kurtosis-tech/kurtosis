---
title: Multi-Phase Runs
sidebar_label: Multi-Phase Runs
---

<!-- TODO Refactor this a bit when we have a 'plan' object -->

Kurtosis environment definitions are encapsulated inside [Starlark scripts][starlark-reference], and these scripts can be bundled into [packages][packages].

Much like Spark, Gradle, Cypress, and Flink, a multi-phase approach is used when Kurtosis runs Starlark:

<!-- TODO Add a dependency phase when we do dependency resolution before interpretation? -->
1. **Interpretation Phase:** The Starlark is uploaded to the Kurtosis engine and the Starlark code is run. Each [function call on the `Plan` object][plan-starlark-reference] adds a step to a plan of instructions to execute, _but the instruction isn't executed yet_.
1. **Validation Phase:** The plan of instructions is validated to ensure port dependencies are referencing existing ports, container images exist, duplicate services aren't being created, etc.
1. **Execution Phase:** The validated plan of instructions is executed, in the order they were defined.

Practically, the user should be aware that:

- Running [a function on the `Plan` object][plan-starlark-reference] does not execute the instruction on-the-spot; it instead adds the instruction to a plan of instructions to execute during the Execution Phase.
- Any value returned by a `Plan` function in Starlark is not the actual value - it is [a future reference that Kurtosis will replace during the Execution Phase when the value actually exists][future-references-reference].

To read about why Kurtosis uses this multi-phase approach, [see here][multi-phase-runs-explanation].

<!---------------- ONLY LINKS BELOW HERE ------------------------->
[starlark-reference]: ./starlark.md
[plan-starlark-reference]: ../starlark-reference/plan.md
[packages]: ./packages.md
[multi-phase-runs-explanation]: ../explanations/why-multi-phase-runs.md
[future-references-reference]: ./future-references.md
