---
title: Idempotent Runs
sidebar_label: Idempotent Runs
---

Idempotent runs refers to Kurtosis' ability to make calls of [`kurtosis run`](../cli-reference/run.md) against an [enclave][enclaves] idempotent, meaning that the [plan](./plan.md) being submitted via `kurtosis run` is a declarative state of how the enclave should look and Kurtosis makes it so regardless of the current state of the enclave. In plain English, this means that Kurtosis will diff the plan being submitted via `kurtosis run` against what already exists in the enclave, and make only the changes necessary to get to the desired state.


This has several uses:

- **Speed:** when you're running a large [Starlark](./starlark.md) script or package and a step near the end has a bug, you don't want to start over from scratch with a fresh enclave and redo all the previous steps. Idempotent runs allows you to simply fix your bug and resubmit, and Kurtosis will skip all the steps that have already been run. 
- **Eternal environments:** eternal environments like shared Dev or Staging are instantiated at their start (Day 0), and then receive a constant updates into the future (Days 1+). In order for these environments to live in Kurtosis, Kurtosis needs to be able to handle Days 1+. Idempotent runs allow this, as you to simply update your Starlark script and Kurtosis updates the environment in the enclave to match.
- **GitOps:** the best DevOps companies in the world use Git to manage changes to environments: each commit updates the environment. This requires idempotency (what happens if the deploy fails for a transient reason? what happens if you need to revert a commit?). Kurtosis' idempotent runs pave the way for a native GitOps experience inside of Kurtosis itself, where the environment infrastructure-as-code is the Starlark itself.

Kurtosis is able to do this because of its [multi-phase approach to running Starlark](./multi-phase-runs.md). Kurtosis constructs an abstract representation of the system you want before running anything (much like Terraform), so Kurtosis can compare the current state of the enclave to the desired state of the enclave and skip any unnecessary changes.

To read in much more detail about how idempotent runs work, see [here](../advanced-concepts/how-do-idempotent-runs-work.md).

<!-------------------------- ONLY LINKS BELOW HERE -------------------------------------->
[enclaves]: ./enclaves.md
