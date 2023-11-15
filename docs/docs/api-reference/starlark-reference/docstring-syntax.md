---
title: Docstring Syntax
sidebar_label: Docstring Syntax
---

If you add a docstring in a specific format to the `run` function inside your package's `main.star` file, you will be able to run your package in [Kurtosis Cloud](https://cloud.kurtosis.com) using a user-friendly form interface.

For example, this is the form produced by Kurtosis for running the Postgres package:

<img src="../../../static/img/api-reference/starlark-reference/postgres-package-configuration-form.png" />

which Kurtosis autogenerates based on [the docstring comment on the package's `main.star` function](https://github.com/kurtosis-tech/postgres-package/blob/main/main.star#L27):

~~~python
def run(
    plan,
    image="postgres:alpine",
    service_name="postgres",
    user="postgres",
    password="MyPassword1!",
    database="postgres",
    config_file_artifact_name="",
    seed_file_artifact_name="",
    extra_configs=[],
    persistent=True,
    launch_adminer=False,
):
    """Launches a Postgresql database instance, optionally seeding it with a SQL file script

    Args:
        image (string): The container image that the Postgres service will be started with
        service_name (string): The name to give the Postgres service
        user (string): The user to create the Postgres database with
        password (string): The password to give to the created user
        database (string): The name of the database to create
        config_file_artifact_name (string): The name of a files artifact that contains a Postgres config file in it
            If not empty, this will be used to configure the Postgres server
        seed_file_artifact_name (string): The name of a files artifact containing seed data
            If not empty, the Postgres server will be populated with the data upon start
        extra_configs (list[string]): Each argument gets passed as a '-c' argument to the Postgres server
        persistent (bool): Whether the data should be persisted. Defaults to True; Note that this isn't supported on multi node k8s cluster as of 2023-10-16
        launch_adminer (bool): Whether to launch adminer which launches a website to inspect postgres database entries. Defaults to False.
    Returns:
        An object containing useful information about the Postgres database running inside the enclave:
        ```
        {
            "database": "postgres",
            "password": "MyPassword1!",
            "port": {
                "application_protocol": "postgresql",
                "number": 5432,
                "transport_protocol": "TCP",
                "wait": "2m0s"
            },
            "service": {
                "hostname": "postgres",
                "ip_address": "172.16.0.4",
                "name": "postgres",
                "ports": {
                    "postgresql": {
                        "application_protocol": "postgresql",
                        "number": 5432,
                        "transport_protocol": "TCP",
                        "wait": "2m0s"
                    }
                }
            },
            "url": "postgresql://postgres:MyPassword1!@postgres/postgres",
            "user": "postgres"
        }
        ```
    """
~~~

The syntax for this docstring is [the Google function docstring syntax](https://google.github.io/styleguide/pyguide.html#383-functions-and-methods), with optional PEP-484 type annotations in parentheses:

```python
def run(plan, required_arg, optional_arg=True, untyped_arg=None):
    """
    This is the `run` function description. It can even spill
    over into multiple lines.

    Args:
        required_arg (string): This is the first argument description. It is required, and is of string type.
            Note that argument descriptions can spill into the second line, so long as they're indented.
        optional_arg (bool): This is the second argument description. It is optional, and it is a boolean-type argument.
        untyped_arg: This argument doesn't have a PEP-484 type annotation.
    """"
```

Several things to note about this docstring syntax:

- Every function MUST have a description before the `Args` block (until [this issue](https://github.com/kurtosis-tech/kurtosis-package-indexer/issues/58) is resolved)
- The description after the `:` (including continuation lines) is rendered in the form using Markdown. For example, entering the following:
  ```
  some_arg (string): Some description about this argument.
      See [this link](https://some-location.com) for more information.
  ```
  would render a link in the description of the argument in the generated form.
- The names of the arguments in the docstring must exactly correspond to the names of the function arguments
- The `plan` argument does not need to be documented; it corresponds to the [`Plan` object](./plan.md) and Kurtosis will inject it automatically
- If an argument has  no default value, it will be marked as required in the Kurtosis-generated form
- If an argument has a default value, it will be marked as optional in the Kurtosis-generated form
- The type of an argument (specified within `()`) is optional

### Types
The values available for use within the `()` to document the type of a function argument are as follows:

Primitives:
- `string`
- `int`
- `bool`

Complex types:
- `list[X]` where `X` is a primitive type
- `dict[X, X]`  where `X` is a primitive type

If a type argument is not specified, Kurtosis will instead generate a JSON field for inputting arbitrary values into the field. The JSON value that the user enters will be transliterated to a Starlark `dict` object.

For example, if the user enters the following JSON in the form field of an argument without a type specified:

```json
{
    "some_key": {
        "inner_key": "Inner key value"
    },
    "some_bool": true
}
```

then the equivalent Starlark that will be passed in for the function parameter will be:

```python
{
    "some_key": {
        "inner_key": "Inner key value",
    },
    "some_bool": True,
}
```
