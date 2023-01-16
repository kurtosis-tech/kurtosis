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
    # The percentage of packets that will be dropped between the two designated subnetworks
    # MANDATORY
    packet_loss_percentage = 50.0,
)
```

:::tip
See [kurtosis.connection][connection-config-prebuilt] for pre-built [ConnectionConfig][connection-config] objects
:::

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
        "path/to/file/1": "files_artifact_1",
        "path/to/file/2": "files_artifact_2",
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

    # Defines the subnetwork in which the service will be started.
    # OPTIONAL (Default: "default")
    subnetwork = "service_subnetwork",
)
```

The `ports` dictionary argument accepts a key value pair, where `key` is a user defined unique port identifier and `value` is a [PortSpec][port-spec] object.

For more info about the `subnetwork` argument, see [Kurtosis subnetworks][subnetworks-reference].

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

`kurtosis.connection.ALLOWED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `0`. It represents a [ConnectionConfig][connection-config] that _allows_ all connection between two subnetworks.

#### `BLOCKED`

`kurtosis.connection.BLOCKED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `100`. It represents a [ConnectionConfig][connection-config] that _blocks_ all connection between two subnetworks.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: #connectionconfig
[service-config]: #serviceconfig
[port-spec]: #portspec

[connection-config-prebuilt]: #connection

[future-references-reference]: ./future-references.md
[subnetworks-reference]: ./subnetworks.md

[starlark-instructions-add-service]: ./starlark-instructions.md#add_service
[starlark-instructions-set-connection]: ./starlark-instructions.md#set_connection
