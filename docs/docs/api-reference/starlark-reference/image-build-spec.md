---
title: ImageBuildSpec
sidebar_label: ImageBuildSpec
---

Kurtosis starts services based on a provided image definition in the `image` arg of [`ServiceConfig`](./service-config.md). You can provide Kurtosis with a published image to use or alternatively, use `ImageBuildSpec` to instruct Kurtosis to build the Docker image the service will be started from. 

`ImageBuildSpec` can be especially useful when developing on a service that needs to be run in an enclave over and over to test latest changes. Kurtosis leverages the Docker's image caching when building images so images aren't rebuilt everytime.

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
Note that `ImageBuildSpec` can only be used in packages and not standalone scripts as it relies on the build context being in the package.
:::