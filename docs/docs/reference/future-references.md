---
title: Future References
sidebar_label: Future References
---

Kurtosis uses [a multi-phase approach][multi-phase-runs-reference] when running [Starlark scripts][starlark-reference].

For the user, the important thing to remember is that any value returned by [a Kurtosis Starlark instruction][starlark-reference] is not the actual value - it is a special string referencing the _future_ value that will exist during the Execution Phase.

For example:

```python
service = add_service(
    "my-service",
    config = struct(
        image = "hello-world",
    )
)
print(service.ip_address)
```

does not in fact print the actual IP address of the service, because the service does not exist during the Interpretation Phase. Instead, `service.ip_address` is a string referencing the future value of the service's IP address:

```
{{kurtosis:my-service.ip_address}}
```

Anywhere this future reference string is used, Kurtosis will slot in the actual value during the Execution Phase. For example, when the `print` statement is executed during the Execution Phase, Kurtosis will replace the future reference with the actual value so that the service's actual IP address gets printed:

```
> print "{{kurtosis:my-service.ip_address}}"
172.19.10.3
```

:::caution
The format of these future reference strings is undefined and subject to change; users should not construct them manually!
:::

All values that are available exclusively during the Execution Phase will be handled in Starlark as future reference strings. This includes:

- Service information
- Files artifact information
- Execution-time values (e.g. values returned by HTTP requests or `exec`ing a command on a container)

<!----------- ONLY LINKS BELOW HERE ----------------------->
[multi-phase-runs-reference]: ./multi-phase-runs.md
[starlark-reference]: ../starlark-reference/index.md
