---
title: Args
sidebar_label: Args
---

Kurtosis [packages][packages-reference] can be parameterized with arguments. Arguments can be passed in via the CLI when running the package.

To make your package take in arguments, first change your `run` function from:

```python
def run(plan):
```

to:

```python
def run(plan, args)
```

Then pass JSON-serialized arg values to `kurtosis run` in the CLI. For example:

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter":"some_value","some_other_param":5}'
```

Kurtosis will automatically JSON-deserialize the JSON string, and then pass it in to the `run` function in Starlark.

The JSON passed in via the command line will be deserialized to a dictionary in Starlark (_not_ a `struct`). So to access the args above, your `main.star` might look like:

```python
def run(plan, args):
    plan.print("some_parameter value: " + args["some_parameter"])
    plan.print("some_other_param value: " + args["some_other_param"])
```

<!------------------------------------- ONLY LINKS BELOW HERE --------------------------------->
[packages-reference]: ./packages.md
