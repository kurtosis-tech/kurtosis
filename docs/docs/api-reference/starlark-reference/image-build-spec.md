---
title: Image Build Spec
sidebar_label: Image Build Spec
---

Kurtosis starts services based on a provided image definition in the `image` arg of `ServiceConfig`. You can provide Kurtosis with a published image to use or alternatively, use `ImageBuildSpec` to instruct Kurtosis to build the Docker image the service will be started from. 

`ImageBuildSpec` can be especially useful when developing on a service that needs to be run in an enclave over and over to test the latest changes to the service. Kurtosis leverages the underlying Docker image cache when building images.

```python
    image = ImageBuildSpec(
        # Name to give built image
        # MANDATORY
        image_name="kurtosistech/example-datastore-server"
        
        # Locator to build context within the Kurtosis package
        # As of now, Kurtosis expects a Dockerfile at the root of the build context
        # MANDATORY
        build_context_dir="./server"
        
        # Stage of image build to target for multi-stage container image
        # OPTIONAL
        target_stage=""
    )
```

:::info
Kurtosis leverages the underlying Docker image cache when building images.
:::