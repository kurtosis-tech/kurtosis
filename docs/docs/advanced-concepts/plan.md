---
title: Plan
sidebar_label: Plan
---

The plan is a representation of what Kurtosis will do inside an enclave. It is central to the [multi-phase run design of Kurtosis][multi-phase-runs]. Plans are built via [Starlark][starlark-reference] by calling [functions on the `Plan` object][plan-starlark-reference] like `add_service`, `remove_service`, or `upload_files`.

You never construct a `Plan` object in Starlark. Instead, the `run` function of your `main.star` should have a variable called `plan`, and Kurtosis will inject the `Plan` object into it. You can then pass the object down to any functions that need it.

For example:

```python
# ------ main.star ---------
some_library = import_module("github.com/some-org/some-library/lib.star")

def run(plan):
    plan.print("Hello, world!")

    some_library.do_something(plan)
```

:::caution
Any value returned by a `Plan` function (e.g. `Plan.add_service`, `Plan.upload_files`) is a [future-reference][future-reference] to the actual value that will only exist at execution time. This means that you cannot run conditionals or manipulations on it in Starlark, at interpretation time! 

Instead, do the manipulation you need at execution time, using something like [`Plan.run_sh`][plan-run-sh] or [`Plan.run_python`][plan-run-python].
:::

<!------------------ ONLY LINKS BELOW HERE -------------------->
[future-reference]: ./future-references.md
[multi-phase-runs]: ./multi-phase-runs.md
[starlark-reference]: ./starlark.md
[plan-starlark-reference]: ../api-reference/starlark-reference/plan.md
[plan-run-sh]: ../api-reference/starlark-reference/plan.md#run_sh
[plan-run-python]: ../api-reference/starlark-reference/plan.md#run_python
