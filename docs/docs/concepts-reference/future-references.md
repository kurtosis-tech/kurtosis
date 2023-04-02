---
title: Future References
sidebar_label: Future References
---

Kurtosis uses [a multi-phase approach][multi-phase-runs-reference] when running [Starlark scripts][starlark-reference].

For the user, the important thing to remember is that any value returned by [the functions on the `Plan` object][plan-starlark-reference] is not the actual value. Rather, it is a special string referencing the _future_ value that will exist during the Execution Phase.

For example, this script...

```python
service = add_service(
    "my-service",
    config = struct(
        image = "hello-world",
    )
)
print("The IP address is " + service.ip_address)
```

...does not in fact print the actual IP address of the [service][services-reference], because the service does not exist during the Interpretation Phase. Instead, `service.ip_address` is a string pointing to the future value of the service's IP address. Practically, that string looks like:

```
{{kurtosis:my-service.ip_address}}
```

Anywhere this future reference string is used, Kurtosis will slot in the actual value during the Execution Phase. For example, when the `print` statement above is executed during the Execution Phase, Kurtosis will replace the future reference with the actual value so that the service's actual IP address gets printed:

```
> print "The IP address is {{kurtosis:my-service.ip_address}}"
The IP address is 172.19.10.3
```

:::caution
The format of these future reference strings is undefined and subject to change. You should not construct them manually!
:::

All values that are available only during the Execution Phase will be handled in Starlark as future reference strings. This includes:

- All resource UUIDs
- Service runtime information (e.g. IP addresses and hostnames)
- Execution-time result values (e.g. the results of [`Plan.request`][plan-request-starlark-reference] or [`Plan.exec`][plan-request-starlark-reference])

<!----------- ONLY LINKS BELOW HERE ----------------------->
[multi-phase-runs-reference]: ./multi-phase-runs.md
[starlark-reference]: ../starlark-reference/introduction.md
[plan-starlark-reference]: ../starlark-reference/plan.md
[services-reference]: ./services.md
[plan-request-starlark-reference]: ../starlark-reference/plan.md#request
[plan-exec-starlark-reference]: ../starlark-reference/plan.md#exec
