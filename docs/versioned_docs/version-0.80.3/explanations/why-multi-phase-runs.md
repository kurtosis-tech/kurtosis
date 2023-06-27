---
title: Why Multi-Phase Runs?
sidebar_label: Why Multi-Phase Runs?
---

Kurtosis runs its [Starlark environment definitions][starlark-reference] in [multiple phases][multi-phase-runs-reference].

This can create pitfalls for users not accustomed to the multi-phase idea. For example, the following code has a bug:

```python
service = add_service(
    "my-service",
    config = struct(
        image = "hello-world",
    )
)

if service.ip_address == "1.2.3.4":
    print("IP address matched")
```

The `if` statement will never evaluate to true because `service.ip_address` is in fact a [future reference string][future-references-reference] that will never equal `1.2.3.4`.

We chose this multi-phase approach despite its complexity because it provides significant advantages over traditional scripting:

- Kurtosis can show the user the plan to execute before any changes are made.
- Kurtosis can validate the entire plan, saving the user from errors like container image typos, service name typos, and port typos before any changes are made.
- Kurtosis can optimize performance (e.g. downloading all container images before anything is executed, which is especially important on timing-sensitive tests).

In the future, the multi-phase approach will also:

- Give the user the power to have new services defined in the plan depend on existing services already running in the enclave.
- Give the user the power to remove and edit parts of the plan they don't like (useful when consuming third-party definitions).
- Allow users to "time-travel" through the plan, discovering what the environment will look like at any point during plan execution.
- Permit compiling a Kurtosis plan down to an idempotent, declarative format (e.g. Helm, Terraform, or Docker Compose).

<!----------------- ONLY LINKS BELOW HERE ----------------->
[starlark-reference]: ../concepts-reference/starlark.md
[multi-phase-runs-reference]: ../concepts-reference/multi-phase-runs.md
[future-references-reference]: ../concepts-reference/future-references.md
