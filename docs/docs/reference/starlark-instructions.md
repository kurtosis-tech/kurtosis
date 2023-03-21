---
title: Starlark Instructions
sidebar_label: Starlark Instructions
sidebar_position: 3
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

### add_service

The `add_service` instruction on the [`plan`][plan-reference] object adds a service to the Kurtosis enclave within which the script executes.

```python
service = plan.add_service(
    # The service name of the service being created.
    # The service name is a reference to the service, which can be used in the future to refer to the service.
    # Service names of active services are unique per enclave.
    # MANDATORY
    service_name = "example-datastore-server-1",

    # The configuration for this service. See the 'ServiceConfig' section of 'Starlark Types' from the sidebar for more information.
    # MANDATORY
    config = service_config,
)
```

For more info about the `config` argument, see [ServiceConfig][starlark-types-service-config]

:::info
See [here][files-artifacts-reference] for more details on files artifacts.
:::

The `add_service` function returns a `service` object that contains service information in the form of [future references][future-references-reference] that can be used later in the script. The `service` struct has:
- A `hostname` property representing [a future reference][future-references-reference] to the service's hostname.
- An `ip_address` property representing [a future reference][future-references-reference] to the service's IP address.
- A `ports` dictionary containing [future reference][future-references-reference] information about each port that the service is listening on.

The value of the `ports` dictionary is an object with three properties, `number`, `transport_protocol` and `application_protocol` (optional), which themselves are [future references][future-references-reference].

Example:
```python
dependency = plan.add_service(
    service_name = "dependency",
    config = ServiceConfig(
        image = "dependency",
        ports = {
            "http": PortSpec(number = 80),
        },
    ),
)

dependency_http_port = dependency.ports["http"]

plan.add_service(
    service_name = "dependant",
    config = ServiceConfig(
        env_vars = {
            "DEPENDENCY_URL": "http://{}:{}".format(dependency.ip_address, dependency_http_port.number),
        },
    ),
)
```

### add_services

The `add_services` instruction on the [`plan`][plan-reference] object adds multiple services all at once.

The main advantage compared to calling [add_service][add-service] multiple times is that one call to `add_services` 
will add multiple services at once. Currently, it can process up to 4 services concurrently, therefore reducing the
total time by a factor of nearly 4.

Similar to `add_service`, it takes a map of service names to service configuration objects as input and returns a map 
of service names to service objects.

```python
all_services = plan.add_services(
    # A map of service_name -> ServiceConfig for all services that needs to be added.
    # See the 'ServiceConfig' section of 'Starlark Types' from the sidebar for more information on this type.
    # MANDATORY
    configs = {
        "example-datastore-server-1": datastore_server_config_1,
        "example-datastore-server-2": datastore_server_config_2,
    }
)
```

:::caution

`add_services` will succeed if and only if all services are successfully added. If any one fails, the entire batch of
services will be rolled back and the instruction will return an execution error.

:::

The number of services being added concurrently is tunable by the `--parallelism` flag of the run command (see more on the [`run`](./cli/run-starlark.md) reference).

### assert

The `assert` on the [`plan`][plan-reference] object instruction fails the Starlark script or package with an execution error if the assertion defined fails.

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

### exec

The `exec` instruction on the [`plan`][plan-reference] object executes commands on a given service as if they were running in a shell on the container.

```python
exec_recipe = ExecRecipe(
    # The actual command to execute. 
    # Each item corresponds to one shell argument, so ["echo", "Hello world"] behaves as if you ran "echo" "Hello world" in the shell.
    # MANDATORY
    command = ["echo", "Hello, world"],
)

result = plan.exec(
    # The recipe that will be run until assert passes.
    # Valid values are of the following types: (ExecRecipe)
    # MANDATORY
    recipe = exec_recipe,

    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "my-service",
)

plan.print(result["output"])
plan.print(result["code"])
```

The instruction returns a `dict` whose values are [future reference][future-references-reference] to the output and exit code of the command. `result["output"]` is a future reference to the output of the command, and `result["code"]` is a future reference to the exit code.

They can be chained to [`assert`][assert] and [`wait`][wait]:

```python
exec_recipe = ExecRecipe(
    service_name = "my_service",
    command = ["echo", "Hello, world"],
)

result = plan.exec(exec_recipe)
plan.assert(result["code"], "==", 0)

plan.wait(exec_recipe, "output", "!=", "Greetings, world")
```

### import_module

The `import_module` function imports the symbols from a Starlark script specified by the given [locator][locators-reference], and requires that the calling Starlark script is part of a [package][packages-reference].

```python
# Import the code to namespaced object
lib = import_module("github.com/foo/bar/src/lib.star")

# Use code from the imported module
lib.some_function()
lib.some_variable
```

NOTE: We chose not to use the normal Starlark `load` primitive due to its lack of namespacing. By default, the symbols imported by `load` are imported to the global namespace of the script that's importing them. We preferred module imports to be namespaced, in the same way that Python does by default with its `import` statement.

### print

`print` on the [`plan`][plan-reference] object will add an instruction to the plan to print the string. When the `print` instruction is executed during the [Execution Phase][multi-phase-runs-reference], [future references][future-references-reference] will be replaced with their execution-time values.

```
plan.print("Any string here")
```

### read_file

The `read_file` function reads the contents of a file specified by the given [locator][locators-reference], and requires that the Starlark script is part of a [package][packages-reference]. `read_file` executes [at interpretation time][multi-phase-runs-reference] so the file contents won't be displayed in the preview.

 ```python
contents = read_file(
    # The Kurtosis locator of the file to read.
    # MANDATORY
    src = "github.com/kurtosis-tech/datastore-army-package/README.md",
)
 ```

### remove_connection

As opposed to `set_connection`, `remove_connection` removes a connection override between two [subnetworks][subnetworks-reference]. The default connection cannot be removed; it can only be updated using [set_connection][set-connection].

```python
remove_connection(
    # The subnetwork connection that will be removed
    # If any of those two subnetworks does not currently have services, this instruction will not do anything.
    # MANDATORY
    subnetworks = ("subnetwork_1", "subnetwork_2"),

)
```

### remove_service

The `remove_service` instruction on the [`plan`][plan-reference] object removes a service from the enclave in which the instruction executes in.

```python
plan.remove_service(
    # The service name of the service to be removed.
    # MANDATORY
    service_name = "my_service",
)
```

### render_templates

`render_templates` on the [`plan`][plan-reference] object combines a template and data to produce a [files artifact][files-artifacts-reference]. Files artifacts can be used with the `files` property in the service config of `add_service`, allowing for reuse of config files across services.

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

### request

The `request` instruction on the [`plan`][plan-reference] object executes either a POST or GET HTTP request, saving its result in a [future references][future-references-reference].

For GET requests:

```python
get_request_recipe = GetHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint?input=data",

    # The extract dictionary takes in key-value pairs where:
    # Key is a way you refer to the extraction later on
    # Value is a 'jq' string that contains logic to extract from response body
    # OPTIONAL
    extract = {
        "extracted-field": ".name.id",
    },
)
get_response = plan.request(
    # The recipe that will be run until assert passes.
    # Valid values are of the following types: (GetHttpRequestRecipe, PostHttpRequestRecipe)
    # MANDATORY
    recipe = get_request_recipe,
    
    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "my_service",
)
plan.print(get_response["body"]) # Prints the body of the request
plan.print(get_response["code"]) # Prints the result code of the request (e.g. 200, 500)
plan.print(get_response["extract.extracted-field"]) # Prints the result of running ".name.id" query, that is saved with key "extracted-field"
```

For POST requests:
```python
post_request_recipe = PostHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint",

    # The content type header of the request (e.g. application/json, text/plain, etc)
    # MANDATORY
    content_type = "text/plain",

    # The body of the request
    # MANDATORY
    body = "text body",

    # The method is GET for this example
    # OPTIONAL (Default: {})
    extract = {},
)
post_response = plan.request(
    recipe = post_request_recipe,
    service_name = "my_service",
)
```

The instruction returns a response, which is a `dict` with following key-value pair; the values are a [future reference][future-references-reference] 
* `response["code"]` - returns the future reference to the `status code` of the response 
* `response["body"]` - returns the future reference to the `body` of the the response
* `response["extract.some-custom-field"]` - it is an optional field and returns the future reference to the value extracted from `body`, which is explained below.

`jq`'s [regular expressions](https://devdocs.io/jq-regular-expressions-pcre/) is used to extract the information from the response `body` and is assigned to a custom field. **The `response["body"]` must be a valid json object for manipulating data using `extractions`**. A valid `response["body"]` can be used for extractions like so:

 ```python
 # Assuming response["body"] looks like {"result": {"foo": ["hello/world/welcome"]}}
post_request_recipe = PostHttpRequestRecipe(
    ...
    extract = {
        "second-element-from-list-head": '.result.foo | .[0] | split ("/") | .[1]' # 
    },
)
response = plan.request(
    recipe = post_request_recipe,
)
# response["extract.second-element-from-list-head"] is "world"
# response["body"] is {"result": {"foo": ["hello/world/welcome"]}}
# response["code"] is 200
``` 

NOTE: In the above example, `response` also has a custom field `extract.second-element-from-list-head` and the value is `world` which is extracted from the `response[body]`.

These fields can be used in conjuction with [`assert`][assert] and [`wait`][wait] instructions, like so:
```python
# Following the example above, response["extract.second-element-from-list-head"] is world
response = plan.request(
    recipe = post_request_recipe,
)

# Assert if the extracted field in the response is world
plan.assert(response["extract.second-element-from-list-head"], "==", "world")

# Make a post request and check if the extracted field in the response is world
plan.wait(post_request_recipe, "extract.second-element-from-list-head", "==", "world")
```

### set_connection

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
        ms = 500
    )
)

set_connection(
    subnetworks = ("subnetworkA", "subnetworkB"),
    config = connection_config
)

```
If serviceA is in subnetworkA and serviceB is in subnetworkB, the effective latency for a TCP request between serviceA and serviceB will be 1000ms = 500ms x 2. This is because the latency is applied to both the request (serviceA -> serviceB) and the response (serviceB -> serviceA)
:::


### store_service_files

`store_service_files` on the [`plan`][plan-reference] object copies files or directories from an existing service in the enclave into a [files artifact][files-artifacts-reference]. This is useful when work produced on one container is needed elsewhere.

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

### update_service

The `update_service` instruction updates an existing service without restarting it. For now, only the [service subnetwork][subnetworks-reference] can be updated live. In this case, the service will be moved to the corresponding subnetwork.

```python
update_service(
    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "example-datastore-server-1",

    # The changes to apply to this service. See the 'UpdateServiceConfig' section of 'Starlark Types' from the sidecar for more information.
    # MANDATORY
    config = update_service_config,
)
```

See [UpdateServiceConfig][starlark-types-update-service-config] for more information on the mandatory `config` argument.

### upload_files

`upload_files` on the [`plan`][plan-reference] object packages the files specified by the [locator][locators-reference] into a [files artifact][files-artifacts-reference] that gets stored inside the enclave. This is particularly useful when a static file needs to be loaded to a service container.

```python
artifact_name = plan.upload_files(
    # The file to upload into a files a files artifact
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

### wait

The `wait` instruction on the [`plan`][plan-reference] object fails the Starlark script or package with an execution error if the assertion does not succeed in a given period of time.

To learn more about the accepted recipe types, please checkout [ExecRecipe][starlark-types-exec-recipe], [GetHttpRequestRecipe][starlark-types-get-http-recipe] or [PostHttpRequestRecipe][starlark-types-post-http-recipe].

If it succeeds, it returns a [future references][future-references-reference] with the last recipe run.


```python
# This fails in runtime if response["code"] != 200 for each request in a 5 minute time span
response = plan.wait(
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
    # OPTIONAL (Default: "15m")
    timeout = "5m",

    # A Service name designating a service that already exists inside the enclave
    # If it does not, a validation error will be thrown
    # MANDATORY
    service_name = "example-datastore-server-1",
)
# If this point of the code is reached, the assertion has passed therefore the print statement will print "200"
plan.print(response["code"])
```

Starlark Standard Libraries
---------------------------

The following Starlark libraries that ship with the `starlark-go` are included
in Kurtosis Starlark by default

1. The Starlark [time](https://github.com/google/starlark-go/blob/master/lib/time/time.go#L18-L52) is a collection of time-related functions
2. The Starlark [json](https://github.com/google/starlark-go/blob/master/lib/json/json.go#L28-L74) module allows you `encode`, `decode` and `indent` JSON
4. The Starlark [struct](https://github.com/google/starlark-go/blob/master/starlarkstruct/struct.go) builtin allows you to create `structs` like the one used in [`add_service`](#add_service)


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[set-connection]: #set_connection
[add-service]: #add_service
[wait]: #wait
[assert]: #assert

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
