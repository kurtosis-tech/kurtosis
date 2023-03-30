---
title: Starlark
sidebar_label: Starlark Reference
slug: /starlark-reference
sidebar_position: 1
---

This page lists out the Kurtosis instructions that are available in Starlark.

**GENERAL NOTE:** In Python, it is very common to name function parameters that are optional. E.g.:

```python
def do_something(required_arg, optional_arg="default_value")
```

In Kurtosis Starlark, all parameters can be referenced by name regardless of whether they are required or not. We do this to allow for ease-of-reading clarity. Mandatory and optional parameters will be indicated in the comment above the field.

Similarly, all function arguments can be provided either positionally or by name. E.g. a function signature of:

```python
def make_pizza(size, topping = "pepperoni")
```

Can be called in any of the following ways:

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
[set-connection]: #set_connection
[add-service]: #add_service
[wait]: #wait
[assert]: #assert
[extract]: #extract

[files-artifacts-reference]: ./files-artifacts.md
[future-references-reference]: ./future-references.md
[packages-reference]: ./packages.md
[locators-reference]: ./locators.md
[multi-phase-runs-reference]: ./multi-phase-runs.md
[plan-reference]: ./plan.md
[subnetworks-reference]: ./subnetworks.md

[starlark-types-connection-config]: ./starlark-types.md#connectionconfig
[starlark-types-service-config]: ./starlark-types.md#serviceconfig
[starlark-types-update-service-config]: ./starlark-types.md#updateserviceconfig
[starlark-types-exec-recipe]: ./starlark-types.md#execrecipe
[starlark-types-post-http-recipe]: ./starlark-types.md#posthttprequestrecipe
[starlark-types-get-http-recipe]: ./starlark-types.md#gethttprequestrecipe
