---
title: Plan
sidebar_label: Plan
---

The plan is a representation of the steps that Kurtosis will execute within an enclave in the future. It is central to the [multi-phase run design of Kurtosis][multi-phase-runs]. Plans are built via [Starlark][starlark-reference] by calling [functions on the `Plan` object][plan-starlark-reference] like `add_service`, `remove_service`, or `upload_files`.

Kurtosis injects the `Plan` object into the `run` method in the `main.star` of your package or your standalone script. The package or script author must ensure that the first argument is an argument called `plan`, and then use the enclave-modifying functions from it. 

The author must pass the `plan` methods down to any other scripts or packages that require enclave-modifying functions.

For example:

```python
def run(plan, args):
    # Plan is passed down to helper functions
    datastore.create_datastore(plan)

def create_datastore(plan):
    plan.add_service(
        name = "datastore-service",
        config = ServiceConfig(
            image = "kurtosistech/example-datastore-server"
        )
    )
```

:::caution
Any value returned by a `plan` function is a [future-reference][future-reference]. This means that you can't run conditionals or Interpretation-time methods like `string.split` on them.
:::

<!------------------ ONLY LINKS BELOW HERE -------------------->
[future-reference]: ./future-references.md
[arguments]: ./packages.md#arguments
[multi-phase-runs]: ./multi-phase-runs.md
[starlark-reference]: ./starlark.md
[plan-starlark-reference]: ../starlark-reference/plan.md
