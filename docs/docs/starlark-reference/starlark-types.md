---
title: Starlark Types
sidebar_label: Starlark Types
sidebar_position: 4
---

This page lists out the Kurtosis types that are available in Starlark.

## Type definitions

### GetHttpRequestRecipe


### PostHttpRequestRecipe


### UniformPacketDelayDistribution


### NormalPacketDelayDistribution


### PortSpec

This `PortSpec` constructor creates a PortSpec object that encapsulates information pertaining to a port.

```python
port_spec = PortSpec(
    # The port number which we want to expose
    # MANDATORY
    number = 3000,

    # Transport protocol for the port (can be either "TCP" or "UDP")
    # OPTIONAL (DEFAULT:"TCP")
    transport_protocol = "TCP",

    # Application protocol for the port that will be displayed in front of URLs containing the port
    # For example:
    #  "http" to get a URL of "http://..."
    #  "postgresql" to get a URL of "postgresql://..."
    # OPTIONAL
    application_protocol = "http",
)
```
The above constructor returns a `PortSpec` object that contains port information in the form of a [future reference][future-references-reference] and can be used with
[add_service][add-service-reference] to create services.

### ReadyConditions

The `ReadyConditions` can be used to execute a readiness check after a service is started to confirm that it is ready to receive connections and traffic 

```python
ready_conditions = ReadyConditions(

    # The recipe that will be used to check service's readiness.
    # Valid values are of the following types: (ExecRecipe, GetHttpRequestRecipe or PostHttpRequestRecipe)
    # MANDATORY
    recipe = GetHttpRequestRecipe(
        port_id = "http",
        endpoint = "/ping",
    ),

    # The `field's value` will be used to do the asssertions. To learn more about available fields, 
    # that can be used for assertions, please refer to exec and request instructions.
    # MANDATORY
    field = "code",

    # The assertion is the comparison operation between value and target_value.
    # Valid values are "==", "!=", ">=", "<=", ">", "<" or "IN" and "NOT_IN" (if target_value is list).
    # MANDATORY
    assertion = "==",

    # The target value that value will be compared against.
    # MANDATORY
    target_value = 200,

    # The interval value is the initial interval suggestion for the command to wait between calls
    # It follows a exponential backoff process, where the i-th backoff interval is rand(0.5, 1.5)*interval*2^i
    # Follows Go "time.Duration" format https://pkg.go.dev/time#ParseDuration
    # OPTIONAL (Default: "1s")
    interval = "1s",

    # The timeout value is the maximum time that the readiness check waits for the assertion to be true
    # Follows Go "time.Duration" format https://pkg.go.dev/time#ParseDuration
    # OPTIONAL (Default: "15m")
    timeout = "5m",
)
```

### ServiceConfig

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
    ready_conditions = ReadyConditions(...)
)
```
The `ports` dictionary argument accepts a key value pair, where `key` is a user defined unique port identifier and `value` is a [PortSpec][port-spec] object.
  
The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a file artifact name. (see [upload_files][upload-files-reference], [render_templates][render-templates-reference] and [store_service_files][store-service-reference] to learn more about on how to create file artifacts)

For more info about the `subnetwork` argument, see [Kurtosis subnetworks][subnetworks-reference].

You can see how to configure the [`ReadyConditions` type here][ready-conditions]. 

### UpdateServiceConfig

The `UpdateServiceConfig` contains the attributes of [ServiceConfig][service-config] that are live-updatable. For now, only the `subnetwork`[subnetworks-reference] attribute of a service can be updated once the service is started.

```python
update_service_config = UpdateServiceConfig(
    # The subnetwork to which the service will be moved.
    # "default" can be used to move the service to the default subnetwork
    # MANDATORY
    subnetwork = "subnetwork_1",
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: #connectionconfig
[service-config]: #serviceconfig
[port-spec]: #portspec
[ready-conditions]: #readyconditions

[connection-config-prebuilt]: #connection

[future-references-reference]: ../reference/future-references.md
[subnetworks-reference]: ../reference/subnetworks.md

[add-service-reference]: ./plan.md#add_service
[set-connection-reference]: ./plan.md#set_connection
[exec-reference]: ./plan.md#exec
[wait-reference]: ./plan.md#wait
[upload-files-reference]: ./plan.md#upload_files
[store-service-reference]: ./plan.md#store_service_files
[render-templates-reference]: ./plan.md#render_templates

