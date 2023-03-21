---
title: Starlark Types
sidebar_label: Starlark Types
sidebar_position: 4
---

This page lists out the Kurtosis types that are available in Starlark.

## Type definitions

### ConnectionConfig

The `ConnectionConfig` is used to configure a connection between two [subnetworks][subnetworks-reference] (see [set_connection][starlark-instructions-set-connection]).

```python
connection_config = ConnectionConfig(
    # Percentage of packet lost each way between subnetworks 
    # OPTIONAL
    # DEFAULT: 0.0
    packet_loss_percentage = 50.0,

    # Amount of delay added to packets each way between subnetworks
    # OPTIONAL: Valid value are UniformPacketDelayDistribution or NormalPacketDelayDistribution
    packet_delay_distribution = UniformPacketDelayDistribution(
        # Delay in ms
        ms = 500 
    ) 
)
```

:::tip
See [kurtosis.connection][connection-config-prebuilt] for pre-built [ConnectionConfig][connection-config] objects
:::

### ExecRecipe

The ExecRecipe can be used to run the `command` on the service (see [exec][starlark-instructions-exec]
or [wait][starlark-instructions-wait])

```python
exec_recipe = ExecRecipe(
    # The actual command to execute. 
    # Each item corresponds to one shell argument, so ["echo", "Hello world"] behaves as if you ran "echo 'Hello World'" in the shell.
    # MANDATORY
    command = ["echo", "Hello, World"],
)
```

### HttpRequestRecipe

The `HttpRequestRecipe` is used to make `HTTP` requests to an endpoint. Currently, we support `GET` and `POST` requests. (see [request][starlark-instructions-request] or [wait][starlark-instructions-wait]).

#### GetHttpRequestRecipe

The `GetHttpRequestRecipe` can be used to make `GET` requests.

```python
get_request_recipe = GetHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint?input=data",

    # The extract dictionary takes in key-value pairs where:
    # Key is a way you refer to the extraction later on
    # Value is a 'jq' string that contains logic to extract from response body
    # To lean more about jq, please visit https://devdocs.io/jq/
    # OPTIONAL
    extract = {
        "extractfield" : ".name.id"
    }
)
```

:::info
Important - `port_id` field accepts user defined ID assinged to a port in service's port map while defininig `ServiceConfig`. For example, we have a service config with following port map:

```
    test-service-config = ServiceConfig(
        ports = {
            // "port_id": port_number
            "http": 5000,
            "grpc": 3000
            ...
        }
        ...
    )
```

The user defined port IDs in above port map are: `http` and `grpc`. These can be passed to create http request recipes (`GET` OR `POST`) such as:

```
    recipe = GetHttpRequestRecipe(
        port_id = "http",
        endpoint = "/ping"
        ...
    )
```

This above recipe when used with `request` or `wait` instruction, will make a `GET` request to a service (the `service_name` field must be passed as an instruction's argument) on port `5000` with the path `/ping`.
:::

#### PostHttpRequestRecipe

The `PostHttpRequestRecipe` can be used to make `POST` requests.

```python
post_request_recipe = PostHttpRequestRecipe(
    # The port ID that is the server port for the request
    # MANDATORY
    port_id = "my_port",

    # The endpoint for the request
    # MANDATORY
    endpoint = "/endpoint",

    # The content type header of the request (e.g. application/json, text/plain, etc)
    # MANDATORY
    content_type = "application/json",

    # The body of the request
    # MANDATORY
    body = "{\"data\": \"this is sample body for POST\"}",
    
    # The extract dictionary takes in key-value pairs where:
    # Key is a way you refer to the extraction later on
    # Value is a 'jq' string that contains logic to extract from response body
    # # To lean more about jq, please visit https://devdocs.io/jq/
    # OPTIONAL
    extract = {
        "extractfield" : ".name.id"
    }
)
```

:::caution

Make sure that the endpoint returns valid JSON response for both POST and GET requests.

:::

### PacketDelayDistribution

The `PacketDelayDistribution` can be used in conjuction with [`ConnectionConfig`][connection-config] to introduce latency between two [`subnetworks`][subnetworks-reference]. See [`set_connection`][starlark-instructions-set-connection] instruction to learn more about its usage.

#### UniformPacketDelayDistribution

The `UniformPacketDelayDistribution` creates a packet delay distribution with constant delay in `ms`

```python

delay  = UniformPacketDelayDistribution(
    # Non-Negative Integer
    # Amount of constant delay added to outgoing packets from the subnetwork
    # MANDATORY
    ms = 1000
)
```

#### NormalPacketDelayDistribution

The `NormalPacketDelayDistribution` can be used to create packet delays that are distributed according to a normal distribution.

```python

delay  = NormalPacketDelayDistribution(
    # Non-Negative Integer
    # Amount of mean delay added to outgoing packets from the subnetwork
    # MANDATORY
    mean_ms = 1000

    # Non-Negative Integer
    # Amount of variance (jitter) added to outgoing packets from the subnetwork
    # MANDATORY
    std_dev_ms = 10
    
    # Non-Negative Float
    # Percentage of correlation observed among packets. It means that the delay observed in next packet
    # will exhibit a corrlation factor of 10.0% with the previous packet. 
    # OPTIONAL
    # DEFAULT = 0.0
    correlation = 10.0
)   
```

### PortSpec

This `PortSpec` constructor creates a PortSpec object that encapsulates information pertaining to a port.

```python
port_spec = PortSpec(
    # The port number which we want to expose
    # MANDATORY
    number = 3000,
    
    # Transport protocol for the port (can be either "TCP" or "UDP")
    # Optional (DEFAULT:"TCP")
    transport_protocol = "TCP",

    # Application protocol for the port
    # Optional
    application_protocol = "http"
)
```
The above constructor returns a `PortSpec` object that contains port information in the form of a [future reference][future-references-reference] and can be used with
[add_service][starlark-instructions-add-service] to create services.

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

The `ServiceConfig` is used to configure a service when it is added to an enclave (see [add_service][starlark-instructions-add-service]).

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
            application_protocol = "http"
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
  
The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a file artifact name. (see [upload_files][starlark-instructions-upload-files], [render_templates][starlark-instructions-render-templates] and [store_service_files][starlark-instructions-store-service-files] to learn more about on how to create file artifacts)

For more info about the `subnetwork` argument, see [Kurtosis subnetworks][subnetworks-reference].

You can see how to configure the [`ReadyConditions` type here][ready-conditions]. 

### UpdateServiceConfig

The `UpdateServiceConfig` contains the attributes of [ServiceConfig][service-config] that are live-updatable. For now, only the `subnetwork`[subnetworks-reference] attribute of a service can be updated once the service is started.

```python
update_service_config = UpdateServiceConfig(
    # The subnetwork to which the service will be moved.
    # "default" can be used to move the service to the default subnetwork
    # MANDATORY
    subnetwork = "subnetwork_1"
)
```

## The global `kurtosis` object

Kurtosis provides "pre-built" values for types that will be broadly used. Those values are provided through the `kurtosis` object. It is available globally and doesn't need to be imported.

### `connection`

#### `ALLOWED`

`kurtosis.connection.ALLOWED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `0` and `packet_delay` set to `PacketDelay(delay_ms=0)`. It represents a [ConnectionConfig][connection-config] that _allows_ all connection between two subnetworks with no delay and packet loss.

#### `BLOCKED`

`kurtosis.connection.BLOCKED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `100` and `packet_delay` set to `PacketDelay(delay_ms=0)`. It represents a [ConnectionConfig][connection-config] that _blocks_ all connection between two subnetworks.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: #connectionconfig
[service-config]: #serviceconfig
[port-spec]: #portspec
[ready-conditions]: #readyconditions

[connection-config-prebuilt]: #connection

[future-references-reference]: ./future-references.md
[subnetworks-reference]: ./subnetworks.md

[starlark-instructions-add-service]: ./starlark-instructions.md#add_service
[starlark-instructions-set-connection]: ./starlark-instructions.md#set_connection
[starlark-instructions-request]: ./starlark-instructions.md#request
[starlark-instructions-wait]: ./starlark-instructions.md#wait
[starlark-instructions-exec]: ./starlark-instructions.md#exec
[starlark-instructions-upload-files]: ./starlark-instructions.md#upload_files
[starlark-instructions-store-service-files]: ./starlark-instructions.md#store_service_files
[starlark-instructions-render-templates]: ./starlark-instructions.md#render_templates

