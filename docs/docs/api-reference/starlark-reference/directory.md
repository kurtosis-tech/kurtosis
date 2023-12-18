---
title: Directory
sidebar_label: Directory
---

The `Directory` constructor creates a `Directory` object that represents a directory inside an existing service (see
the [ServiceConfig][service-config] object).
Directory object can be either a files artifact or a persistent directory, depending on the arguments passed to the 
constructor.

```python
file_artifact_directory = Directory(
    artifact_name=files_artifact_1,
)
# Or:
persistent_directory = Directory(
    persistent_key="data-directory"
)
```

A directory composed of a files artifact will be automatically provisioned with the given files artifact content. In 
the above example, `files_artifact_1` is a files artifact name. (see [upload_files][upload-files-reference], 
[render_templates][render-templates-reference] and [store_service_files][store-service-reference] to learn more about 
on how to create file artifacts). 

A persistent directory, as its name indicates, persists over service updates and restarts. It is uniquely identified 
by its `persistent_key` and the service ID on which it is being used (a persistent directory cannot be shared across
multiple services). When it is first created, it will be empty. The service can write anything in it. When the service 
gets updated, the data in it persists. It is particularly useful for a service's data directory, logs directory, etc.

A persistent directory of a specific size can be created using the `size` field; the value supplied is in megabytes

```python
persistent_directory = Directory(
    size = 1000
)
```

The default size of a persistent directory is `1Gb`. Note the size attribute is ignored on Docker due to Docker limitations.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[render-templates-reference]: ./plan.md#render_templates
[service-config]: ./service-config.md
[store-service-reference]: ./plan.md#store_service_files
[upload-files-reference]: ./plan.md#upload_files
