---
title: ImageSpec
sidebar_label: ImageSpec
---

The `ImageSpec` object constructor allows for providing detailed information about a container image for use in the [`ServiceConfig.image`](./service-config.md#image) property. It is most commonly used when pulling an image from a non-Dockerhub container repository.

Signature
---------

```
ImageSpec(
    name, 
    registry = "http://hub.docker.com", 
    username = None, 
    password = None,
)
```

| Property | Description |
| --- | --- |
| **name**<br/>_string_ | The name of the container image that should be used. |
| **registry**<br/>_string_ | The registry that the container should be pulled from. |
| **username**<br/>_string_ | The username that will be used for connecting to the registry when pulling the image. |
| **password**<br/>_string_ | The password that will be used for connecting to the registry when pulling the image. |

Examples
--------
Starting a service called `my-service` from `my.registry.io` using a custom username and password:

```python
plan.add_service(
    name = "my-service",
    config = ServiceConfig(
        image = ImageSpec(
            name = "my.registry.io/some-org/some-image",
            registry = "http://my.registry.io/",
            username = "registry-user",
            password = "SomePassword1!",
        )
    )
)
```
