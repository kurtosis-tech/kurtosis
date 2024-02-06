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
    # Service names of active services are unique per enclave and needs to be formatted according to RFC 1035. 
    # Specifically, 1-63 lowercase alphanumeric characters with dashes and cannot start or end with dashes.
    # Also service names have to start with a lowercase alphabet.
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
        image=ImageBuildSpec(image_name="dependant", build_context_dir="./"),
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

verify
------

The `verify` instruction throws an [Execution phase error][multi-phase-runs-reference] if the defined assertion fails.

```python
plan.verify(
    # The value currently being verified.
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

plan.verify(
    # Value can also be a runtime value derived from a `get_value` call
    value = response["body"],
    assertion = "==",
    target_value = 200,
)
```

:::caution

Verifications are typed, so running

```python
plan.verify(
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

They can be chained to [`verify`][verify] and [`wait`][wait]:

```python
exec_recipe = ExecRecipe(
    command = ["echo", "Hello, world"],
)

result = plan.exec(service_name="my_service", recipe=exec_recipe)
plan.verify(result["code"], "==", 0)

plan.wait(service_name="my_service", recipe=exec_recipe, field="output", assertion="!=", target_value="Greetings, world")
```

print
-----

The `print` instruction will print a value during [the Execution phase][multi-phase-runs-reference]. When the `print` instruction is executed during the Execution Phase, [future references][future-references-reference] will be replaced with their execution-time values.

```
plan.print("Any string here")
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

**Returns**: a [future reference][future-references-reference] resolving to a `string` representing the name of a [files artifact][files-artifacts-reference].

**Args**:
- `config`: a dictionary with the following keys and values:
  - **keys**: `string`s representing the filepaths to be produced within the returned files artifact
  - **values**: `struct`s with the following root level keys:
    - `template`: a string with representing the template in [Go template format](https://pkg.go.dev/text/template#pkg-overview)
    - `data`: a `struct` or `dict` type, with keys matching the variables used in the template, and values matching the intended replacement values.

**Examples**:

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

**See also**:
- [add-service]
- [add-services]
- [service-config]

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

These fields can be used in conjunction with [`verify`][verify] and [`wait`][wait] instructions, like so:
```python
# Following the example above, response["extract.second-element-from-list-head"] is world
post_response = plan.request(
    service_name = "my_service",
    recipe = post_request_recipe,
)

# Assert if the extracted field in the response is world
plan.verify(response["extract.second-element-from-list-head"], "==", "world")

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

run_python
----------

The `run_python` instruction executes a one-time execution task. It runs the Python script specified by the mandatory field `run` on an image specified by the optional `image` field.

```python
    result = plan.run_python(
        # The Python script to execute as a string
        # This will get executed via '/bin/sh -c "python /tmp/python/main.py"'.
        # Where `/tmp/python/main.py` is path on the temporary container;
        # on which the script is written before it gets run
        # MANDATORY
        run = """
    import requests
    response = requests.get("docs.kurtosis.com")
    print(response.status_code)      
    """,

        # Arguments to be passed t o the Python script defined in `run`
        # OPTIONAL (Default: [])
        args = [
            some_other_service.ports["http"].url,
        ],

        # Packages that the Python script requires which will be installed via `pip`
        # OPTIONAL (default: [])
        packages = [
            "selenium",
            "requests",
        ],
    
        # Image the Python script will be run on
        # OPTIONAL (Default: python:3.11-alpine)
        image = "python:3.11-alpine",

        # A mapping of path_on_task_where_contents_will_be_mounted -> files_artifact_id_to_mount
        # For more information about file artifacts, see below.
        # CAUTION: duplicate paths to files or directories to be mounted is not supported, and it will fail
        # OPTIONAL (Default: {})
        files = {
            "/path/to/file/1": files_artifact_1,
            "/path/to/file/2": files_artifact_2,
        },

        # A list of filepaths to store inside files artifacts after the run_python finishes
        # Entries in the list can either be a string containing the path to store, or a
        # StoreSpec object that can optionally name the files artifact.
        # CAUTION: Both the paths in `src` and the files artifact names must be unique!
        # OPTIONAL (Default: [])
        store = [
            # EXAMPLE: Creates a files artifact named `kurtosis_txt` containing the `kurtosis.txt` file
            StoreSpec(src = "/src/kurtosis.txt", name = "kurtosis_txt"),
            
            # EXAMPLE: Creates a files artifact with an automatically-generated name containing `genesis.json`
            StoreSpec(src = "/genesis.json"),

            # EXAMPLE: Creates a files artifact with an automatically-generated name containing `address.json`
            # This is just syntactic sugar for:
            # StoreSpec(src = "/coinbase/address.json")
            "/coinbase/address.json"
        ],

        # The time to allow for the command to complete. If the Python script takes longer than this,
        # Kurtosis will kill the script and mark it as failed.
        # You may specify a custom wait timeout duration or disable the feature entirely.
        # You may specify a custom wait timeout duration with a string:
        #  wait = "2m"
        # Or, you can disable this feature by setting the value to None:
        #  wait = None
        # The feature is enabled by default with a default timeout of 180s
        # OPTIONAL (Default: "180s")
        wait="180s"
    )

    plan.print(result.code)  # returns the future reference to the exit code
    plan.print(result.output) # returns the future reference to the output
    plan.print(result.file_artifacts) # returns the file artifact names that can be referenced later
```

The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a [file artifact][files-artifacts-reference] name.

The instruction returns a `struct` with [future references][future-references-reference] to the output and exit code of the Python script, alongside with future-reference to the file artifact names that were generated.
* `result.output` is a future reference to the output of the command
* `result.code` is a future reference to the exit code
* `result.files_artifacts` is a future reference to the names of the file artifacts that were generated and can be used by the `files` property of `ServiceConfig` or `run_sh` instruction. An example is shown below:-

run_sh
-------------

The `run_sh` instruction executes a one-time execution task. It runs the bash command specified by the mandatory field `run` on an image specified by the optional `image` field.

```python
    result = plan.run_sh(
        # The command to run, as a string
        # This will get executed via 'sh -c "$COMMAND"'.
        # For example: 'sh -c "mkdir -p kurtosis && echo $(ls)"'
        # MANDATORY
        run = "mkdir -p kurtosis && echo $(ls)",

        # Image the command will be run on
        # OPTIONAL (Default: badouralix/curl-jq)
        image = "badouralix/curl-jq",

        # Defines environment variables that should be set inside the Docker container running the task.
        # OPTIONAL (Default: {})
        env_vars = {
            "VAR_1": "VALUE_1",
            "VAR_2": "VALUE_2",
        },

        # A mapping of path_on_task_where_contents_will_be_mounted -> files_artifact_id_to_mount
        # For more information about file artifacts, see below.
        # CAUTION: duplicate paths to files or directories to be mounted is not supported, and it will fail
        # OPTIONAL (Default: {})
        files = {
            "/path/to/file/1": files_artifact_1,
            "/path/to/file/2": files_artifact_2,
        },

        # A list of filepaths to store inside files artifacts after the run_sh finishes
        # Entries in the list can either be a string containing the path to store, or a
        # StoreSpec object that can optionally name the files artifact.
        # CAUTION: Both the paths in `src` and the files artifact names must be unique!
        # OPTIONAL (Default: [])
        store = [
            # EXAMPLE: Creates a files artifact named `kurtosis_txt` containing the `kurtosis.txt` file
            StoreSpec(src = "/src/kurtosis.txt", name = "kurtosis_txt"),
            
            # EXAMPLE: Creates a files artifact with an automatically-generated name containing `genesis.json`
            StoreSpec(src = "/genesis.json"),

            # EXAMPLE: Creates a files artifact with an automatically-generated name containing `address.json`
            # This is just syntactic sugar for:
            # StoreSpec(src = "/coinbase/address.json")
            "/coinbase/address.json"
        ],

        # The time to allow for the command to complete. If the command takes longer than this,
        # Kurtosis will kill the command and mark it as failed.
        # You may specify a custom wait timeout duration or disable the feature entirely.
        # You may specify a custom wait timeout duration with a string:
        #  wait = "2m"
        # Or, you can disable this feature by setting the value to None:
        #  wait = None
        # The feature is enabled by default with a default timeout of 180s
        # OPTIONAL (Default: "180s")
        wait="180s"
    )

    plan.print(result.code)  # returns the future reference to the code
    plan.print(result.output) # returns the future reference to the output
    plan.print(result.file_artifacts) # returns the file artifact names that can be referenced later
```

The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a [file artifact][files-artifacts-reference] name.

The instruction returns a `struct` with [future references][future-references-reference] to the output and exit code of the command, alongside with future-reference to the file artifact names that were generated. 
   * `result.output` is a future reference to the output of the command
   * `result.code` is a future reference to the exit code
   * `result.files_artifacts` is a future reference to the names of the file artifacts that were generated and can be used by the `files` property of `ServiceConfig` or `run_sh` instruction. An example is shown below:-

```python

    result = plan.run_sh(
        run = "mkdir -p task && cd task && echo kurtosis > test.txt",
        store = [
            "/task",
            # using '*' will only copy the contents of the parent directory and not the directory itself to file artifact
            # in this case, only test.txt will be stored and task directory will be ignored
            "/task/*", 
        ],
        ...
    )

    plan.print(result.files_artifacts) # prints ["blue_moon", "green_planet"]
    
    # blue_moon is name of the file artifact that contains task directory
    # green_planet is the name of the file artifact that contains test.txt file

    service_one = plan.add_service(
        ..., 
        config=ServiceConfig(
            name="service_one", 
            files={"/src": results.file_artifacts[0]}, # copies the directory task into service_one 
        )
    ) # the path to the file will look like: /src/task/test.txt

    service_two = plan.add_service(
        ..., 
        config=ServiceConfig(
            name="service_two", 
            files={"/src": results.file_artifacts[1]}, # copies the file test.txt into service_two
        ),
    ) # the path to the file will look like: /src/test.txt
```

start_service
-------------

The `start_service` instruction restarts a service that was stopped temporarily by [`stop_service`][stop-service].

```python
plan.start_service(
    # The service name of the service to be restarted.
    # MANDATORY
    name = "my_service",
)
```

stop_service
------------

The `stop_service` instruction stops a service temporarily.  The container ends but its configuration stays around so it can be restarted quickly using [`start_service`][start-service].

```python
plan.stop_service(
    # The service name of the service to be stopped.
    # MANDATORY
    name = "my_service",
)
```

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


upload_files
------------

The `upload_files` instruction packages the files specified by the [locator][locators-reference] into a [files artifact][files-artifacts-reference] that gets stored inside the enclave. This is particularly useful when a static file needs to be loaded to a service container. 

```python
artifact_name = plan.upload_files(
    # The file to upload into a files artifact
    # Must be any GitHub URL without the '/blob/main' part.
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

The `wait` instruction fails the Starlark script or package with an execution error if the provided [verification][verify] does not succeed within a given period of time. 

If the assertion succeeds, `wait` returns the result of the given Recipe  - i.e. the same output as [`plan.request`][request] or [`plan.exec`][exec].

This instruction is best used for asserting the system has reached a desired state, e.g. in testing. To wait until a service is ready, you are better off using automatic port availability waiting via [`PortSpec.wait`][starlark-types-port-spec] or [`ServiceConfig.ready_conditions`][ready-condition], as these will short-circuit a parallel [`add_services`][add-services] call if they fail.

To learn more about the accepted recipe types, please see [`ExecRecipe`][starlark-types-exec-recipe], [`GetHttpRequestRecipe`][starlark-types-get-http-recipe] or [`PostHttpRequestRecipe`][starlark-types-post-http-recipe].


```python
# This fails in runtime if response["code"] != 200 for each request in a 5 minute time span
recipe_result = plan.wait(
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

# The assertion has passed, so we can use `recipe_result` just like the result of `plan.request` or `plan.exec`
plan.print(recipe_result["code"])
```


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[add-service]: #add_service
[add-services]: #add_services
[verify]: #verify
[extract]: #extract
[exec]: #exec
[request]: #request
[start-service]: #start_service
[stop-service]: #stop_service
[wait]: #wait

[cli-run-reference]: ../../cli-reference/run.md

[files-artifacts-reference]: ../../advanced-concepts/files-artifacts.md
[future-references-reference]: ../../advanced-concepts/future-references.md
[packages-reference]: ../../advanced-concepts/packages.md
[locators-reference]: ../../advanced-concepts/locators.md
[multi-phase-runs-reference]: ../../advanced-concepts/multi-phase-runs.md
[ready-condition]: ./ready-condition.md
[service-config]: ./service-config.md

[starlark-types-service-config]: ./service-config.md
[starlark-types-exec-recipe]: ./exec-recipe.md
[starlark-types-post-http-recipe]: ./post-http-request-recipe.md
[starlark-types-get-http-recipe]: ./get-http-request-recipe.md
[service-starlark-reference]: ./service.md
[starlark-types-port-spec]: ./port-spec.md
