---
title: Args
sidebar_label: Args
---

Kurtosis [packages][packages-reference] can be parameterized with arguments. Arguments can be passed in via the CLI when running the package.

To make your package take in arguments, first add extra parameters besides the `plan` to your package's `run` function. 

From this...

```python
def run(plan):
```

...to this...

```python
def run(plan, some_parameter, some_other_parameter="Default value"):
```

Then pass JSON-serialized arg values to `kurtosis run` in the CLI, with each key in the JSON being a parameter to the `run` function. 

For example:

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter": 5, "some_other_parameter": "New value"}'
```

Kurtosis deserializes the JSON, with each key treated as a separate parameter passed to the `run` function in Starlark. 

This is the equivalent to the following Starlark:

```python
run(plan, some_parameter = 5, some_other_parameter = "New value")
```

:::info
By default, Kurtosis deserializes JSON objects (anything in `{}`) as dictionaries in Starlark. However, sometimes you need to pass a `struct` as a parameter instead. 

To have Kurtosis deserialize a JSON object as a `struct` instead of a dictionary, simply add `"_kurtosis_parser": "struct"` to the object.

For example, this command...

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter": {"_kurtosis_parser": "struct", "some_property": "Property value"}}'
```

...is equivalent to this Starlark:

```python
run(plan, some_parameter = struct(some_property = "Property value"))
```

:::

<!------------------------------------- ONLY LINKS BELOW HERE --------------------------------->
[packages-reference]: ./packages.md
