---
title: Directory
sidebar_label: Directory
---

The `Directory` constructor creates a `Directory` object that represents a directory inside an existing service (see
the [ServiceConfig][service-config] object).
Directory object can be a files artifacts list or a persistent directory, depending on the arguments passed to the 
constructor.

```python
file_artifact_directory = Directory(
    artifact_names=[files_artifact_1, files_artifact_2] 
)
# Or:
persistent_directory = Directory(
    persistent_key="data-directory"
)
```

A directory composed of one or many files artifacts will be automatically provisioned with the given files artifacts content. In 
the above examples, `files_artifact_1` and `files_artifact_2` are files artifact names. (see [upload_files][upload-files-reference], 
[render_templates][render-templates-reference] and [store_service_files][store-service-reference] to learn more about 
on how to create file artifacts). 

:::warning
Take into account when using multiple file artifacts, like the example, if both files artifact contains a file or folder with the same name this will end up overwritten.

A persistent directory, as its name indicates, persists over service updates and restarts. It is uniquely identified 
by its `persistent_key` (a persistent directory cannot be shared across
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
