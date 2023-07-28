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
def run(plan, some_other_param, some_parameter="Default value")
```

Then pass JSON-serialized arg values to `kurtosis run` in the CLI. For example:

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter":"some_value","some_other_param":5}'
```

Kurtosis will automatically JSON-deserialize the JSON string, and then pass it in to the `run` function in Starlark.

<!------------------------------------- ONLY LINKS BELOW HERE --------------------------------->
[packages-reference]: ./packages.md
