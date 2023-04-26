---
title: ReadyCondition
sidebar_label: ReadyCondition
---

The `ReadyCondition` can be used to execute a readiness check after a service is started to confirm that it is ready to receive connections and traffic.

It is flexible, and you can configure it in order to check service readiness in many ways, but if you just want to check if your service's ports are open (to consider it ready) and these can receive incoming calls, we suggest to use the `wait` field in the [ServiceConfig][service-config] type. And if you need one-off wait not attached to the [ServiceConfig][service-config] you suggest you to use the [wait][wait] instruction

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

Let us show you how to use it with an easy example:
```python
def run(plan):
    # we define the recipe first
    get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)

    # then the ready conditions using the ReadyCondition type which contain the recipe already created
    ready_conditions_config = ReadyCondition(
        recipe = get_recipe,
        field = "code",
        assertion = "==",
        target_value = 200,
        interval = "10s",
        timeout = "200s",
    )

    # we set the ready conditions in the ServiceConfig 
    service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		},
        ready_conditions= ready_conditions_config,
	)

    # finally we execute the add_service instruction using all the pre-configured data
    plan.add_service(name = "web-server", config = service_config)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->

[service-config]: ./service-config.md
[port-spec]: ./port-spec.md
[wait]: ./plan.md#wait