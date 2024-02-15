---
title: PostHttpRequestRecipe
sidebar_label: PostHttpRequestRecipe
---

The `PostHttpRequestRecipe` can be used to make `POST` requests to an endpoint.

```python
post_request_recipe = PostHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint",

    # The content type header of the request (e.g. application/json, text/plain, etc)
    # OPTIONAL (DEFAULT:"application/json")
    content_type = "application/json",

    # The body of the request
    # OPTIONAL (DEFAULT:"")
    body = "{\"data\": \"this is sample body for POST\"}",
    
    # The extract dictionary takes in key-value pairs where:
    # Key is a way you refer to the extraction later on
    # Value is a 'jq' string that contains logic to extract from response body
    # # To lean more about jq, please visit https://devdocs.io/jq/
    # OPTIONAL (DEFAULT:{})
    extract = {
        "extractfield" : ".name.id",
    },
)
```

:::caution

Make sure that the endpoint returns valid JSON response for both POST and GET requests.

:::

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
