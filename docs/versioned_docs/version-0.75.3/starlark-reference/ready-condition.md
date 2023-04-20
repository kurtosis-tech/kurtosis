---
title: ReadyCondition
sidebar_label: ReadyCondition
---

The `ReadyCondition` can be used to execute a readiness check after a service is started to confirm that it is ready to receive connections and traffic 

```python
ready_conditions = ReadyCondition(

    # The recipe that will be used to check service's readiness.
    # Valid values are of the following types: (ExecRecipe, GetHttpRequestRecipe or PostHttpRequestRecipe)
    # MANDATORY
    recipe = GetHttpRequestRecipe(
        port_id = "http",
        endpoint = "/ping",
    ),

    # The `field's value` will be used to do the asssertions. To learn more about available fields, 
    # that can be used for assertions, please refer to exec and request instructions.
    # MANDATORY
    field = "code",

    # The assertion is the comparison operation between value and target_value.
    # Valid values are "==", "!=", ">=", "<=", ">", "<" or "IN" and "NOT_IN" (if target_value is list).
    # MANDATORY
    assertion = "==",

    # The target value that value will be compared against.
    # MANDATORY
    target_value = 200,

    # The interval value is the initial interval suggestion for the command to wait between calls
    # It follows a exponential backoff process, where the i-th backoff interval is rand(0.5, 1.5)*interval*2^i
    # Follows Go "time.Duration" format https://pkg.go.dev/time#ParseDuration
    # OPTIONAL (Default: "1s")
    interval = "1s",

    # The timeout value is the maximum time that the readiness check waits for the assertion to be true
    # Follows Go "time.Duration" format https://pkg.go.dev/time#ParseDuration
    # OPTIONAL (Default: "15m")
    timeout = "5m",
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
