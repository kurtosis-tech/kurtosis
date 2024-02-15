---
title: ExecRecipe
sidebar_label: ExecRecipe
---

The ExecRecipe can be used to run the `command` on the service (see [exec][exec-reference]
or [wait][wait-reference])

```python
exec_recipe = ExecRecipe(
    # The actual command to execute. 
    # Each item corresponds to one shell argument, so ["echo", "Hello world"] behaves as if you ran "echo 'Hello World'" in the shell.
    # MANDATORY
    command = ["echo", "Hello, World"],
        
    # The extract dictionary can be used for filtering specific parts of a response
    # assigning that output to a key-value pair, where the key is the reference 
    # variable and the value is the specific output. 
    # 
    # Specifically: the key is the way you refer to the extraction later on and
    # the value is a 'jq' string that contains logic to extract parts from response 
    # body that you get from the exec_recipe used
    # 
    # To lean more about jq, please visit https://devdocs.io/jq/
    # OPTIONAL (DEFAULT:{})
    extract = {
        "extractfield" : ".name.id",
    },
)
```

:::tip
If you are trying to run a complex `command` with `|`, you should prefix the command with `/bin/sh -c` and wrap the actual command in a string; for example: `command = ["echo", "a", "|", "grep a"]` should
be rewritten as `command = ["/bin/sh", "-c", "echo a | grep a"]`. Not doing so makes everything after the `echo` as args of that command, instead of following the behavior you would expect from a shell.
:::

:::tip
If the executed command returns a proper `JSON` formatted data structure, it's necessary to pass the output through `jq`'s `fromjson` function to enable `jq` to parse the input.
For more information on `jq`'s built-in methods, plese refer to `jq`'s documentation. The following is an example of how to parse the json formatted output using `jq` syntax:

Example:
```
def run(plan, args={}):
    plan.add_service(
        name = "service",
        config = ServiceConfig(
            image = "alpine",
            entrypoint = ["/bin/sh", "-c", "sleep infinity"],
        )
    )
    cmd = ''' echo '{"key": "value"}' '''
    result = plan.exec(
        service_name = "service",
        recipe = ExecRecipe(
            command = ["/bin/sh", "-c", cmd],
            extract = {
                "example_reference_key": "fromjson | .key"   # <----- Notice the use of `fromjson`
            }
        ),
    )
    plan.print(result["output"])
    plan.print(result["extract.example_reference_key"])
```

will output:
```
> print msg="{{kurtosis:1f60460f3eee4036af01b41fc2ecddc0:output.runtime_value}}"
{"key": "value"}


> print msg="{{kurtosis:1f60460f3eee4036af01b41fc2ecddc0:extract.example_reference_key.runtime_value}}"
value
```

:::


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[exec-reference]: ./plan.md#exec
[wait-reference]: ./plan.md#wait
