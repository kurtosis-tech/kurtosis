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
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[exec-reference]: ./plan.md#exec
[wait-reference]: ./plan.md#wait
