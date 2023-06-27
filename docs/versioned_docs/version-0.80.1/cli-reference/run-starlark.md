---
title: run
sidebar_label: run
slug: /run-starlark
---

Kurtosis can be used to run a Starlark script or a [runnable package](../concepts-reference/packages.md) in an enclave. 

A single Starlark script can be ran with:

```bash
kurtosis run script.star
```

Adding the `--dry-run` flag will print the changes without executing them. 

A [Kurtosis package](../concepts-reference/packages.md) on your local machine can be run with:

```bash
kurtosis run /path/to/package/on/your/machine
```

A [runnable Kurtosis package](../concepts-reference/packages.md) published to GitHub can be run like so:

```bash
kurtosis run github.com/package-author/package-repo
```

:::tip
If you want to run a non-main branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

Arguments can be provided to a Kurtosis package (either local or from GitHub) by passing a JSON-serialized object with args argument, which is the second positional argument you pass to `kurtosis run` like:

```bash
# Local package
kurtosis run /path/to/package/on/your/machine '{"company":"Kurtosis"}'

# GitHub package
kurtosis run github.com/package-author/package-repo '{"company":"Kurtosis"}'
```

:::info
If the flag `--main-function-name` is set, the JSON-serialized object will be used for passing the function arguments.

For example, if the main function signature (inside this file github.com/my-org/my-package/src/entry.star) has this shape:
```python
# the plan object will automatically be injected if the first argument name is 'plan'
def my_main_function(plan, first_argument, second_argument, their_argument):
    # your code
```

It can be called like this:
```bash
# you don't have to pass the plan object as an argument because it will automatically be injected by default if the first argument name is 'plan'
kurtosis run main.star '{"first_argument": "Foo", "second_argument": "Bar", "their_argument": {"first-key:"first-value", "second-key":"second-value"}}'  --main-file src/entry.star --main-function-name my_main_function
```

THIS IS A TEMPORARY OPTION AND IT WILL BE REMOVED SOON!!!
:::

This command has options available to customize its execution:

1. The `--dry-run` flag can be used to print the changes proposed by the script without executing them
1. The `--parallelism` flag can be used to specify to what degree of parallelism certain commands can be run. For example: if the script contains an [`add_services`][add-services-reference] instruction and is run with `--parallelism 100`, up to 100 services will be run at one time.
1. The `--enclave` flag can be used to instruct Kurtosis to run the script inside the specified enclave or create a new enclave (with the given enclave [identifier](../concepts-reference/resource-identifier.md)) if one does not exist. If this flag is not used, Kurtosis will create a new enclave with an auto-generated name, and run the script or package inside it.
1. The `--with-subnetworks` flag can be used to enable [subnetwork capabilties](../concepts-reference/subnetworks.md) within the specified enclave that the script or package is instructed to run within. This flag is false by default.
1. The `--verbosity` flag can be used to set the verbosity of the command output. The options include `BRIEF`, `DETAILED`, or `EXECUTABLE`. If unset, this flag defaults to `BRIEF` for a concise and explicit output. Use `DETAILED` to display the exhaustive list of arguments for each command. Meanwhile, `EXECUTABLE` will generate executable Starlark instructions. 
1. The `--main-file` flag can be used to set the main file filepath, the "main" file is a file for the main method (i.e. the package's entrypoint) which will be executed first; the filepath has to be relative to the package's root. The default value is 'main.star'. This flag is only used for running packages. Example: if your main file is located in a path like this `github.com/my-org/my-package/src/internal/my-file.star` you should set `src/internal/my-file.star` as the relative path.
1. The `--main-function-name` flag can be used to set the main function name, which will be executed first as the entrypoint of the package or the module. The default value is 'run'.

Example of using setting the --main-function-name flag

For example, to run the `start_node` function in a `main.star` file, simple use:
```bash
kurtosis run main.star --main-function-name start_node
```

Where start-node is a function defined in `main.star` as so:
```python
# main.star code
def start_node(plan,args):
    # your code
```

<!--------------------------------------- ONLY LINKS BELOW HERE -------------------------------->
[add-services-reference]: ../starlark-reference/plan.md#add_services
