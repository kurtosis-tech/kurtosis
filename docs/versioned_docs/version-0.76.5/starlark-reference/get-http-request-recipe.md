---
title: GetHttpRequestRecipe
sidebar_label: GetHttpRequestRecipe
---

The `GetHttpRequestRecipe` can be used to make `GET` requests to an endpoint, filter for the specific part of the response you care about, and assign that specific output to a key for later use. This can be useful for writing assertions, for example (i.e. validating the response you end up receiving looks the way you expect/intended).

```python
get_request_recipe = GetHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint?input=data",

    # The extract dictionary can be used for filtering specific parts of a HTTP GET
    # request and assigning that output to a key-value pair, where the key is the
    # reference variable and the value is the specific output. 
    # 
    # Specifcally: the key is the way you refer to the extraction later on and
    # the value is a 'jq' string that contains logic to extract parts from response 
    # body that you get from the HTTP GET request.
    # 
    # To lean more about jq, please visit https://devdocs.io/jq/
    # OPTIONAL (DEFAULT:{})
    extract = {
        "extractfield" : ".name.id",
    },
)
```

:::info
Important - the `port_id` field accepts user-defined port IDs that are assigned to a port in a service's port map, using `ServiceConfig`. For example, if our service's `ServiceConfig` has the following port mappings:

```
    test-service-config = ServiceConfig(
        ports = {
            // "port_id": port_number
            "http": 5000,
            "grpc": 3000,
            ...
        },
        ...
    )
```

The user-defined port IDs in the above `ServiceConfig` are: `http` and `grpc`. Both of these user-defined port IDs can therefore be used to create http request recipes (`GET` OR `POST`), such as:

```
    recipe = GetHttpRequestRecipe(
        port_id = "http",
        service_name = "service-using-test-service-config",
        endpoint = "/ping",
        ...
    )
```

The above recipe, when used with `request` or `wait` instruction, will make a `GET` request to a service (the `service_name` field must be passed as an instruction's argument) on port `5000` with the path `/ping`.
:::

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
