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
    # Specifcally: the key is the way you refer to the extraction later on and
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

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[exec-reference]: ./plan.md#exec
[wait-reference]: ./plan.md#wait
