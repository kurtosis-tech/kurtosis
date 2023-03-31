---
title: Plan
sidebar_label: Plan
---

The plan is a representation of the manipulations that Kurtosis should execute within an enclave. It is central to the [multi-phase run design of Kurtosis][multi-phase-runs]. Plans are built via [Starlark][starlark-reference] by calling [functions on the `Plan` object][plan-starlark-reference] like `add_service`, `remove_service`, or `upload_files`.

Kurtosis injects the `Plan` object into the `run` method in the `main.star` of your package or your standalone script. The package or script author must ensure that the first argument is an argument called `plan`, and then use the enclave-modifying functions from it. The author also must pass the `plan` methods down to any other scripts or packages that require enclave-modifying functions.

Here's an example :-

Imagine you have a `kurtosis.yml` that looks like
```yaml
name: "github.com/test-author/test-package"
```

Further with a `main.star` at the root of the package that looks like
```py
datastore = import_module("github.com/test-author/test-package/lib/datastore.star")

def run(plan):
    datastore.create_datastore(plan)
```

and the `lib/datastore.star` looks like
```py
def create_datastore(plan):
    plan.add_service(
        name = "datastore-service",
        config = ServiceConfig(
            image = "kurtosistech/example-datastore-server"
        )
    )
```

To accept [arguments][arguments] in the `run` function, pass them as the second parameter like so

```py
def run(plan, args):
    pass
```

:::caution
Any value returned by a `plan` function is a [future-reference][future-reference]. This means that you can't run conditionals or interpretation time methods like `string.split` on it.
:::

<!------------------ ONLY LINKS BELOW HERE -------------------->
[future-reference]: ./future-references.md
[arguments]: ./packages.md#arguments
[multi-phase-runs]: ./multi-phase-runs.md
[starlark-reference]: ./starlark.md
[plan-starlark-reference]: ../starlark-reference/plan.md
