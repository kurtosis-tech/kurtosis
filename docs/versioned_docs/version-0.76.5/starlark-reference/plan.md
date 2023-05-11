---
title: Plan
sidebar_label: Plan
---

The `Plan` object is an object representing the steps that Kurtosis will take inside the enclave during [the Execution phase][multi-phase-runs-reference]. 

Kurtosis injects a `Plan` object into the `run` function in the `main.star` of your Starlark script. Kurtosis relies on the first argument of your `run` function being named `plan` (lowercase); all your Starlark scripts must follow this convention.

To use the `Plan` object in inner functions, simply pass the variable down.

Note that the function calls listed here merely add a step to the plan. They do _not_ run the actual execution. Per Kurtosis' [multi-phase run design][multi-phase-runs-reference], this will only happen during the Execution phase. Therefore, all plan functions will return [future references][future-references-reference].

add_service
-----------

The `add_service` instruction adds a service to the Kurtosis enclave within which the script executes, and returns a [`Service`][service-starlark-reference] object containing information about the newly-added service.

```python
# Returns a Service object (see the Service page in the sidebar)
service = plan.add_service(
    # The service name of the service being created.
    # The service name is a reference to the service, which can be used in the future to refer to the service.
    # Service names of active services are unique per enclave.
    # MANDATORY
    name = "example-datastore-server-1",

    # The configuration for this service, as specified via a ServiceConfig object (see the ServiceConfig page in the sidebar)
    # MANDATORY
    config = service_config,
)
```

For detailed information about the parameters the `config` argument accepts, see [ServiceConfig][starlark-types-service-config].

For detailed information about what `add_service` returns, see [Service][service-starlark-reference].

Example:

```python
dependency = plan.add_service(
    name = "dependency",
    config = ServiceConfig(
        image = "dependency",
        ports = {
            "http": PortSpec(number = 80),
        },
    ),
)

dependency_http_port = dependency.ports["http"]

plan.add_service(
    name = "dependant",
    config = ServiceConfig(
        env_vars = {
            "DEPENDENCY_URL": "http://{}:{}".format(dependency.ip_address, dependency_http_port.number),
        },
    ),
)
```

add_services
------------

The `add_services` instruction behaves like `add_service`, but adds the services in parallel.

The default parallelism is 4, but this can be increased using [the `--parallelism` flag of the `run` CLI command][cli-run-reference].

`add_services` takes a dictionary of service names -> [`ServiceConfig`][starlark-types-service-config] objects as input, and returns a dictionary 
of service names -> [`Service`][service-starlark-reference] objects.

```python
all_services = plan.add_services(
    # A map of service_name -> ServiceConfig for all services that needs to be added.
    # See the 'ServiceConfig' page in the sidebar for more information on this type.
    # MANDATORY
    configs = {
        "example-datastore-server-1": datastore_server_config_1,
        "example-datastore-server-2": datastore_server_config_2,
    },
)
```

For detailed information about the `ServiceConfig` object, see [here][starlark-types-service-config].

For detailed information about the `Service` objects that `add_services`, see [Service][service-starlark-reference].


:::caution

`add_services` will succeed if and only if all services are successfully added. If any one fails (perhaps due to timeouts a ready condition failing), the entire batch of
services will be rolled back and the instruction will return an execution error.

:::

assert
------

The `assert` instruction throws an [Execution phase error][multi-phase-runs-reference] if the defined assertion fails.

```python
plan.assert(
    # The value currently being asserted.
    # MANDATORY
    value = "test1",

    # The assertion is the comparison operation between value and target_value.
    # Valid values are "==", "!=", ">=", "<=", ">", "<" or "IN" and "NOT_IN" (if target_value is list).
    # MANDATORY
    assertion = "==",

    # The target value that value will be compared against.
    # MANDATORY
    target_value = "test2",
) # This fails in runtime given that "test1" == "test2" is false

plan.assert(
    # Value can also be a runtime value derived from a `get_value` call
    value = response["body"],
    assertion = "==",
    target_value = 200,
)
```

:::caution

Asserts are typed, so running

```python
plan.assert(
    value = "0",
    assertion = "==",
    target_value = 0,
)
```

Will fail. If needed, you can use the `extract` feature to parse the types of your outputs.
:::

exec
----

The `exec` instruction executes a command on a container as if they were running in a shell on the container.

```python
exec_recipe = ExecRecipe(
    # The actual command to execute. 
    # Each item corresponds to one shell argument, so ["echo", "Hello world"] behaves as if you ran "echo" "Hello world" in the shell.
    # MANDATORY
    command = ["echo", "Hello, world"],
)

result = plan.exec(
    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "my-service",
    
    # The recipe that will determine the exec to be performed.
    # Valid values are of the following types: (ExecRecipe)
    # MANDATORY
    recipe = exec_recipe,
    
    # If the recipe returns a code that does not belong on this list, this instruction will fail.
    # OPTIONAL (Defaults to [0])
    acceptable_codes = [0, 0], # Here both 0 and 1 are valid codes that we want to accept and not fail the instruction
    
    # If False, instruction will never fail based on code (acceptable_codes will be ignored).
    # You can chain this call with assert to check codes after request is done.
    # OPTIONAL (Defaults to False)
    skip_code_check = False,
)

plan.print(result["output"])
plan.print(result["code"])
```

The instruction returns a `dict` whose values are [future reference][future-references-reference] to the output and exit code of the command. `result["output"]` is a future reference to the output of the command, and `result["code"]` is a future reference to the exit code.

They can be chained to [`assert`][assert] and [`wait`][wait]:

```python
exec_recipe = ExecRecipe(
    command = ["echo", "Hello, world"],
)

result = plan.exec(service_name="my_service", recipe=exec_recipe)
plan.assert(result["code"], "==", 0)

plan.wait(service_name="my_service", recipe=exec_recipe, field="output", assertion="!=", target_value="Greetings, world")
```

print
-----

The `print` instruction will print a value during [the Execution phase][multi-phase-runs-reference]. When the `print` instruction is executed during the Execution Phase, [future references][future-references-reference] will be replaced with their execution-time values.

```
plan.print("Any string here")
```


remove_connection
-----------------

As opposed to `set_connection`, `remove_connection` removes a connection override between two [subnetworks][subnetworks-reference]. The default connection cannot be removed; it can only be updated using [set_connection][set-connection].

```python
remove_connection(
    # The subnetwork connection that will be removed
    # If any of those two subnetworks does not currently have services, this instruction will not do anything.
    # MANDATORY
    subnetworks = ("subnetwork_1", "subnetwork_2"),

)
```

remove_service
--------------

The `remove_service` instruction removes a service from the enclave in which the instruction executes in.

```python
plan.remove_service(
    # The service name of the service to be removed.
    # MANDATORY
    name = "my_service",
)
```

render_templates
----------------

The `render_templates` instruction combines a template and data to produce a [files artifact][files-artifacts-reference]. Files artifacts can be used with the `files` property of the `ServiceConfig` object, allowing for reuse of config files across services.

```python
# Example data to slot into the template
template_data = {
    "Name" : "Stranger",
    "Answer": 6,
    "Numbers": [1, 2, 3],
    "UnixTimeStamp": 1257894000,
    "LargeFloat": 1231231243.43,
    "Alive": True,
}

artifact_name = plan.render_templates(
    # A dictionary where:
    #  - Each key is a filepath that will be produced inside the output files artifact
    #  - Each value is the template + data required to produce the filepath
    # Multiple filepaths can be specified to produce a files artifact with multiple files inside.
    # MANDATORY
    config = {
        "/foo/bar/output.txt": struct(
            # The template to render, which should be formatted in Go template format:
            #   https://pkg.go.dev/text/template#pkg-overview
            # MANDATORY
            template="Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}",

            # The data to slot into the template, can be a struct or a dict
            # The keys should exactly match the keys in the template.
            # MANDATORY
            data=template_data,
        ),
    },

    # The name to give the files artifact that will be produced.
    # If not specified, it will be auto-generated.
    # OPTIONAL
    name = "my-artifact",
)
```

The return value is a [future reference][future-references-reference] to the name of the [files artifact][files-artifacts-reference] that was generated, which can be used with the `files` property of the service config of the `add_service` command.

request
-------

The `request` instruction executes either a POST or GET HTTP request, saving its result in a [future references][future-references-reference].

To make a GET or POST request, simply set the `recipe` field to use the specified [GetHttpRequestRecipe][starlark-types-get-http-recipe] or the [PostHttpRequestRecipe][starlark-types-post-http-recipe]. 

```python
http_response = plan.request(
    # A service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "my_service",
    
    # The recipe that will determine the request to be performed.
    # Valid values are of the following types: (GetHttpRequestRecipe, PostHttpRequestRecipe)
    # MANDATORY
    recipe = request_recipe,
    
    # If the recipe returns a code that does not belong on this list, this instruction will fail.
    # OPTIONAL (Defaults to [200, 201, ...])
    acceptable_codes = [200, 500], # Here both 200 and 500 are valid codes that we want to accept and not fail the instruction
    
    # If False, instruction will never fail based on code (acceptable_codes will be ignored).
    # You can chain this call with assert to check codes after request is done.
    # OPTIONAL (defaults to False)
    skip_code_check = false,
)
plan.print(get_response["body"]) # Prints the body of the request
plan.print(get_response["code"]) # Prints the result code of the request (e.g. 200, 500)
plan.print(get_response["extract.extracted-field"]) # Prints the result of running ".name.id" query, that is saved with key "extracted-field"
```

The instruction returns a response, which is a `dict` with following key-value pair; the values are a [future reference][future-references-reference] 
* `response["code"]` - returns the future reference to the `status code` of the response 
* `response["body"]` - returns the future reference to the `body` of the the response
* `response["extract.some-custom-field"]` - it is an optional field and returns the future reference to the value extracted from `body`, which is explained below.

#### extract

`jq`'s [regular expressions](https://stedolan.github.io/jq/manual/) is used to extract the information from the response `body` and is assigned to a custom field. **The `response["body"]` must be a valid json object for manipulating data using `extractions`**. A valid `response["body"]` can be used for extractions. See below for an example of how this can be done for the [PostHttpRequestRecipe][starlark-types-post-http-recipe]:

 ```python
 # Assuming response["body"] looks like {"result": {"foo": ["hello/world/welcome"]}}
post_request_recipe = PostHttpRequestRecipe(
    ...
    extract = {
        "second-element-from-list-head": '.result.foo | .[0] | split ("/") | .[1]',
    },
)
post_response = plan.request(
    service_name = "my_service",
    recipe = post_request_recipe,
)
# response["extract.second-element-from-list-head"] is "world"
# response["body"] is {"result": {"foo": ["hello/world/welcome"]}}
# response["code"] is 200
``` 

NOTE: In the above example, `response` also has a custom field `extract.second-element-from-list-head` and the value is `world` which is extracted from the `response[body]`.

These fields can be used in conjunction with [`assert`][assert] and [`wait`][wait] instructions, like so:
```python
# Following the example above, response["extract.second-element-from-list-head"] is world
post_response = plan.request(
    service_name = "my_service",
    recipe = post_request_recipe,
)

# Assert if the extracted field in the response is world
plan.assert(response["extract.second-element-from-list-head"], "==", "world")

# Make a post request and check if the extracted field in the response is world
plan.wait(service_name="my_service", recipe=post_request_recipe, field="extract.second-element-from-list-head", assertion="==", target_value="world")
```

NOTE: `jq` returns a typed output that translates into the correspondent Starlark type. You can cast it using `jq` to match
your desired output type:

```python
# Assuming response["body"] looks like {"url": "posts/1"}}
post_request_recipe = PostHttpRequestRecipe(
    ...
    extract = {
        "post-number": '.url | split ("/") | .[1]',
        "post-number-as-int": '.url | split ("/") | .[1] | tonumber',
    },
)
response = plan.request(
    service_name = "my_service",
    recipe = post_request_recipe,
)
# response["extract.post-number"] is "1" (starlark.String)
# response["extract.post-number-as-int"] is 1 (starlark.Int)
```

For more details see [ `jq`'s builtin operators and functions](https://stedolan.github.io/jq/manual/#Builtinoperatorsandfunctions)

set_connection
--------------

Kurtosis uses a *default connection* to configure networking for any created subnetwork.
The `set_connection` can be used for two purposes:

1. Used with the `subnetworks` argument, it will override the default connection between the two specified [subnetworks][subnetworks-reference].

```python
set_connection(
    # The subnetwork connection that will be be overridden
    # OPTIONAL: See 2. below
    subnetworks = ("subnetwork_1", "subnetwork_2"),

    # The configuration for this connection. See the 'ConnectionConfig' section of 'Starlark Types' from the sidecar for more information.
    # MANDATORY
    config = connection_config,
)
```

2. Used with only the `config` argument, it will update the *default connection*.

:::caution

Doing so will _immediately_ affect all subnetwork connections that were not previously overridden.

:::

```python
set_connection(
    # The configuration for this connection. See the 'ConnectionConfig' section of 'Starlark Types' from the sidecar for more information.
    # MANDATORY
    config = connection_config,
)
```

See [ConnectionConfig][starlark-types-connection-config] for more information on the mandatory `config` argument.

:::important

Say we are overriding a connection between two subnetworks, as shown below:

```python

connection_config = ConnectionConfig(
    packet_delay_distribution = UniformPacketDelayDistribution(
        ms = 500,
    ),
)

set_connection(
    subnetworks = ("subnetworkA", "subnetworkB"),
    config = connection_config,
)

```
If serviceA is in subnetworkA and serviceB is in subnetworkB, the effective latency for a TCP request between serviceA and serviceB will be 1000ms = 500ms x 2. This is because the latency is applied to both the request (serviceA -> serviceB) and the response (serviceB -> serviceA)
:::


store_service_files
-------------------

The `store_service_files` instruction copies files or directories from an existing service in the enclave into a [files artifact][files-artifacts-reference]. This is useful when work produced on one container is needed elsewhere.

```python
artifact_name = plan.store_service_files(
    # The service name of a preexisting service from which the file will be copied.
    # MANDATORY
    service_name = "example-service-name",

    # The path on the service's container that will be copied into a files artifact.
    # MANDATORY
    src = "/tmp/foo",

    # The name to give the files artifact that will be produced.
    # If not specified, it will be auto-generated.
    # OPTIONAL
    name = "my-favorite-artifact-name",
)
```

The return value is a [future reference][future-references-reference] to the name of the [files artifact][files-artifacts-reference] that was generated, which can be used with the `files` property of the service config of the `add_service` command.

update_service
--------------

The `update_service` instruction updates an existing service without restarting it. For now, only the [service subnetwork][subnetworks-reference] can be updated live. In this case, the service will be moved to the corresponding subnetwork.

```python
update_service(
    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    name = "example-datastore-server-1",

    # The changes to apply to this service. See the 'UpdateServiceConfig' section of 'Starlark Types' from the sidecar for more information.
    # MANDATORY
    config = update_service_config,
)
```

See [UpdateServiceConfig][starlark-types-update-service-config] for more information on the mandatory `config` argument.

upload_files
------------

`upload_files` instruction packages the files specified by the [locator][locators-reference] into a [files artifact][files-artifacts-reference] that gets stored inside the enclave. This is particularly useful when a static file needs to be loaded to a service container.

```python
artifact_name = plan.upload_files(
    # The file to upload into a files artifact
    # Must be a Kurtosis locator.
    # MANDATORY
    src = "github.com/foo/bar/static/example.txt",

    # The name to give the files artifact that will be produced.
    # If not specified, it will be auto-generated.
    # OPTIONAL
    name = "my-artifact",
)
```

The return value is a [future reference][future-references-reference] to the name of the [files artifact][files-artifacts-reference] that was generated, which can be used with the `files` property of the service config of the `add_service` command.

wait
----

The `wait` instruction fails the Starlark script or package with an execution error if the provided [assertion][assert] does not succeed within a given period of time. If the assertion succeeds, `wait` returns a [future references][future-references-reference] with the result the last run of the assertion.

This instruction is best used for asserting the system has reached a desired state, e.g. in testing. To wait until a service is ready, you are better off using automatic port availability waiting via [`PortSpec.wait`][starlark-types-port-spec] or [`ServiceConfig.ready_conditions`][starlark-types-update-service-config], as these will short-circuit a parallel [`add_services`][add-services] call if they fail.

To learn more about the accepted recipe types, please see [`ExecRecipe`][starlark-types-exec-recipe], [`GetHttpRequestRecipe`][starlark-types-get-http-recipe] or [`PostHttpRequestRecipe`][starlark-types-post-http-recipe].


```python
# This fails in runtime if response["code"] != 200 for each request in a 5 minute time span
response = plan.wait(
    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "example-datastore-server-1",

    # The recipe that will be run until assert passes.
    # Valid values are of the following types: (ExecRecipe, GetHttpRequestRecipe, PostHttpRequestRecipe)
    # MANDATORY
    recipe = recipe,

    # Wait will use the response's field to do the asssertions. To learn more about available fields, 
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

    # The timeout value is the maximum time that the command waits for the assertion to be true
    # Follows Go "time.Duration" format https://pkg.go.dev/time#ParseDuration
    # OPTIONAL (Default: "10s")
    timeout = "5m",
)
# If this point of the code is reached, the assertion has passed therefore the print statement will print "200"
plan.print(response["code"])
```


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[set-connection]: #set_connection
[add-service]: #add_service
[add-services]: #add_services
[wait]: #wait
[assert]: #assert
[extract]: #extract

[cli-run-reference]: ../cli-reference/run-starlark.md

[files-artifacts-reference]: ../concepts-reference/files-artifacts.md
[future-references-reference]: ../concepts-reference/future-references.md
[packages-reference]: ../concepts-reference/packages.md
[locators-reference]: ../concepts-reference/locators.md
[multi-phase-runs-reference]: ../concepts-reference/multi-phase-runs.md
[ready-condition]: ./ready-condition.md
[service-config]: ./service-config.md
[subnetworks-reference]: ../concepts-reference/subnetworks.md

[starlark-types-connection-config]: ./connection-config.md
[starlark-types-service-config]: ./service-config.md
[starlark-types-update-service-config]: ./update-service-config.md
[starlark-types-exec-recipe]: ./exec-recipe.md
[starlark-types-post-http-recipe]: ./post-http-request-recipe.md
[starlark-types-get-http-recipe]: ./get-http-request-recipe.md
[service-starlark-reference]: ./service.md
[starlark-types-port-spec]: ./port-spec.md
