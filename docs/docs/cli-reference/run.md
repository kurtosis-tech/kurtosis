---
title: run
sidebar_label: run
slug: /run
---

Kurtosis can be used to run a Starlark script or a [runnable package](../advanced-concepts/packages.md) in an enclave.

A single Starlark script can be ran with:

```bash
kurtosis run script.star
```

Adding the `--dry-run` flag will print the changes without executing them.

A [Kurtosis package](../advanced-concepts/packages.md) on your local machine can be run with:

```bash
kurtosis run /path/to/package/on/your/machine
```

A [Kurtosis package](../advanced-concepts/packages.md) published to GitHub can be run like so:

```bash
kurtosis run github.com/package-author/package-repo
```

:::tip
If you want to run a non-main branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

### Arguments

Package behaviour can be customized by passing in JSON/YAML-serialized arguments when calling `kurtosis run`.

For example, if your package's `run` function looks like this...

```python
def run(plan, some_parameter, some_other_parameter="Default value"):
```

...then you can pass in values for `some_parameter` and `some_other_parameter` like so:

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter": 5, "some_other_parameter": "New value"}'
```

Kurtosis deserializes the JSON, with each key treated as a separate parameter passed to the `run` function in Starlark.

This is the equivalent to the following Starlark:

```python
run(plan, some_parameter = 5, some_other_parameter = "New value")
```

:::info
By default, Kurtosis deserializes JSON objects (anything in `{}`) as dictionaries in Starlark. However, sometimes you need to pass a `struct` as a parameter instead.

To have Kurtosis deserialize a JSON object as a `struct` instead of a dictionary, simply add `"_kurtosis_parser": "struct"` to the object.

For example, this command...

```bash
kurtosis run github.com/USERNAME/REPO '{"some_parameter": {"_kurtosis_parser": "struct", "some_property": "Property value"}}'
```

...is equivalent to this Starlark:

```python
run(plan, some_parameter = struct(some_property = "Property value"))
```
:::

### Extra Configuration

`kurtosis run` has additional flags that can further modify its behaviour:

1. The `--args-file` flag can be used to send in a YAML/JSON file, from a local file through the filepath or from remote using the URL, as an argument to the Kurtosis Package. Note that if you pass in package arguments as CLI arguments and via the flag, the CLI arguments will be the one used.
   For example:
   ```bash
   kurtosis run github.com/ethpandaops/ethereum-package --args-file "devnet-5.yaml"
   ```
   or
   ```bash
   kurtosis run github.com/ethpandaops/ethereum-package --args-file "https://www.myhost.com/devnet-5.json"
   ```

1. The `--dry-run` flag can be used to print the changes proposed by the script without executing them
1. The `--parallelism` flag can be used to specify to what degree of parallelism certain commands can be run. For example: if the script contains an [`add_services`][add-services-reference] instruction and is run with `--parallelism 100`, up to 100 services will be run at one time.
1. The `--enclave` flag can be used to instruct Kurtosis to run the script inside the specified enclave or create a new enclave (with the given enclave [identifier](../advanced-concepts/resource-identifier.md)) if one does not exist. If this flag is not used, Kurtosis will create a new enclave with an auto-generated name, and run the script or package inside it.
1. The `--verbosity` flag can be used to set the verbosity of the command output. The options include `BRIEF`, `DETAILED`, or `EXECUTABLE`. If unset, this flag defaults to `BRIEF` for a concise and explicit output. Use `DETAILED` to display the exhaustive list of arguments for each command and instruction execution time. Meanwhile, `EXECUTABLE` will generate executable Starlark instructions.
1. The `--main-function-name` flag can be used to set the name of Starlark function inside the package that `kurtosis run` will call. The default value is `run`, meaning Starlark will look for a function called `run` in the file defined by the `--main-file` flag (which defaults to `main.star`). Regardless of the function, Kurtosis expects the main function to have a parameter called `plan` into which Kurtosis will inject [the Kurtosis plan](../advanced-concepts/plan.md).

   For example:

   To run the `start_node` function in a `main.star` file, simple use:
   ```bash
   kurtosis run main.star --main-function-name start_node
   ```

   Where `start_node` is a function defined in `main.star` like so:
   ```python
   # --------------- main.star --------------------
   def start_node(plan, args):
       # your code
   ```
1. The `--main-file` flag sets the main file in which Kurtosis looks for the main function defined via the `--main-function-name` flag. This can be thought of as the entrypoint file. This flag takes a filepath **relative to the package's root**, and defaults to `main.star`. For example, if your package is `github.com/my-org/my-package` but your main file is located in subdirectories like `github.com/my-org/my-package/src/internal/my-file.star`, you should set this flag like `--main-file src/internal/my-file.star`.

   Example of using the `--main-function-name` flag

   For example, to run the `start_node` function in a `main.star` file, simple use:
   ```bash
   kurtosis run main.star --main-function-name start_node
   ```

   Where `start_node` is a function defined in `main.star` like so:

   ```python
   # main.star code
   def start_node(plan,args):
       # your code
   ```
1. The `--production` flag can be used to make sure services restart in case of failure (default behavior is not restart)

1. The `--no-connect` flag can be used to disable user services port forwarding (default behavior is to forward the ports)

1. The `--image-download` flag can be used to configure the download behavior for a given run. When set to `missing`, Kurtosis will only download the latest image tag if the image does not already exist locally (irrespective of the tag of the locally cached image). When set to `always`, Kurtosis will always check and download the latest image tag, even if the image exists locally.

1. The `--experimental` flag can be used to enable experimental or incubating features. Please reach out to Kurtosis team if you wish to try any of those.


<!--------------------------------------- ONLY LINKS BELOW HERE -------------------------------->
[add-services-reference]: ../api-reference/starlark-reference/plan.md#add_services
