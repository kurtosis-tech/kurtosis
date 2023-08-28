---
id: index
title: Starlark Introduction
sidebar_label: Introduction
slug: /starlark-reference
sidebar_position: 1
---

This section details the Kurtosis Starlark DSL used to manipulate the contents of enclaves. Feel free to use the [official Kurtosis Starlark VS Code extension][vscode-plugin] when writing Starlark with VSCode for features like syntax highlighting, method signature suggestions, hover preview for functions, and auto-completion for Kurtosis custom types.

Parameter Naming Convention
---------------------------

In Python, it is very common to name function parameters that are optional. E.g.:

```python
def do_something(required_arg, optional_arg="default_value")
```

In Kurtosis Starlark, all parameters can be referenced by name regardless of whether they are required or not. We do this to allow for ease-of-reading clarity. Mandatory and optional parameters will be indicated in the comment above the field. Example:

```python
def make_pizza(
    # If true, the crust will be thin; if false, the crust will be regular
    # MANDATORY
    thin_crust = True,
)
```

Similarly, all function arguments can be provided either positionally or by name.

For example, this function...

```python
def make_pizza(size, topping = "pepperoni")
```

...can be called in any of the following ways:

```python
# 1. Only the required argument filled, positionally
make_pizza("16cm")

# 2. Only the required argument filled, by name 
make_pizza(size = "16cm")

# 3. Both arguments filled, positionally
make_pizza("16cm", "mushroom")

# 4. Both arguments filled, mixing position and name
make_pizza("16cm", topping = "mushroom")

# 5. Both arguments filled, by name
make_pizza(size = "16cm", topping = "mushroom")
```

We recommend the last style (naming both positional and optional args), for reading clarity.


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
