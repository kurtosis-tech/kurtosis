---
title: ServiceConfig
sidebar_label: ServiceConfig
---

The `ServiceConfig` is used to configure a service when it is added to an enclave (see [add_service][add-service-reference]).

```python
config = ServiceConfig(
    # The name of the container image that Kurtosis should use when creating the service’s container.
    # MANDATORY
    image = "kurtosistech/example-datastore-server",

    # The ports that the container should listen on, identified by a user-friendly ID that can be used to select the port again in the future.
    # If no ports are provided, no ports will be exposed on the host machine, unless there is an EXPOSE in the Dockerfile
    # OPTIONAL (Default: {})
    ports = {
        "grpc": PortSpec(
            # The port number which we want to expose
            # MANDATORY
            number = 3000,

            # Transport protocol for the port (can be either "TCP" or "UDP")
            # Optional (DEFAULT:"TCP")
            transport_protocol = "TCP",

            # Application protocol for the port
            # Optional
            application_protocol = "http",
        ),
    },

    # A mapping of path_on_container_where_contents_will_be_mounted -> files_artifact_id_to_mount
    # For more info on what a files artifact is, see below
    # OPTIONAL (Default: {})
    files = {
        "path/to/file/1": files_artifact_1,
        "path/to/file/2": files_artifact_2,
    },

    # The ENTRYPOINT statement hardcoded in a container image's Dockerfile might not be suitable for your needs.
    # This field allows you to override the ENTRYPOINT when the container starts.
    # OPTIONAL (Default: [])
    entrypoint = [
        "bash",
    ],

    # The CMD statement hardcoded in a container image's Dockerfile might not be suitable for your needs.
    # This field allows you to override the CMD when the container starts.
    # OPTIONAL (Default: [])
    cmd = [
        "-c",
        "sleep 99",
    ],

    # Defines environment variables that should be set inside the Docker container running the service. 
    # This can be necessary for starting containers from Docker images you don’t control, as they’ll often be parameterized with environment variables.
    # OPTIONAL (Default: {})
    env_vars = {
        "VAR_1": "VALUE_1",
        "VAR_2": "VALUE_2",
    },

    # ENTRYPOINT, CMD, and ENV variables sometimes need to refer to the container's own IP address. 
    # If this placeholder string is referenced inside the 'entrypoint', 'cmd', or 'env_vars' properties, the Kurtosis engine will replace it at launch time
    # with the container's actual IP address.
    # OPTIONAL (Default: "KURTOSIS_IP_ADDR_PLACEHOLDER")
    private_ip_address_placeholder = "KURTOSIS_IP_ADDRESS_PLACEHOLDER",

    # The maximum amount of CPUs the service can use, in millicpu/millicore.
    # OPTIONAL (Default: no limit)
    cpu_allocation = 1000,

    # The maximum amount of memory, in megabytes, the service can use.
    # OPTIONAL (Default: no limit)
    memory_allocation = 1024,

    # Defines the subnetwork in which the service will be started.
    # OPTIONAL (Default: "default")
    subnetwork = "service_subnetwork",
    
    # This field can be used to check the service's readiness after this is started
    # to confirm that it is ready to receive connections and traffic
    # OPTIONAL (Default: no ready conditions)
    ready_conditions = ReadyCondition(...)
)
```
The `ports` dictionary argument accepts a key value pair, where `key` is a user defined unique port identifier and `value` is a [PortSpec][port-spec] object.
  
The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a file artifact name. (see [upload_files][upload-files-reference], [render_templates][render-templates-reference] and [store_service_files][store-service-reference] to learn more about on how to create file artifacts)

For more info about the `subnetwork` argument, see [Kurtosis subnetworks][subnetworks-reference].

You can see how to configure the [`ReadyCondition` type here][ready-condition]. 

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[add-service-reference]: ./plan.md#add_service
[port-spec]: ./port-spec.md
[upload-files-reference]: ./plan.md#upload_files
[render-templates-reference]: ./plan.md#render_templates
[store-service-reference]: ./plan.md#store_service_files
[subnetworks-reference]: ../concepts-reference/subnetworks.md
[ready-condition]: ./ready-condition.md
