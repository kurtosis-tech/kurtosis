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

    # A mapping of path_on_container_where_contents_will_be_mounted -> Directory object or file artifact name
    # For more info on what a Directory object is, see below
    # 
    # OPTIONAL (Default: {})
    files = {
        "path/to/files/artifact_1/": files_artifact_1,
        "path/to/files/artifact_2/": Directory(
            artifact_name=files_artifact_2,
        ),
        "path/to/persistent/directory/": Directory(
            persistent_key="data-directory",
        ),
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
    max_cpu = 1000,

    # The mimimum amout of CPUs the service must have, in millicpu/millicore.
    # CAUTION: This is only available for Kubernetes, and will be ignored for Docker.
    # OPTIONAL (Default: no limit)
    min_cpu = 500,

    # The maximum amount of memory, in megabytes, the service can use.
    # OPTIONAL (Default: no limit)
    max_memory = 1024,

    # The minimum amount of memory, in megabytes, the service must have.
    # CAUTION: This is only available for Kubernetes, and will be ignored for Docker.
    # OPTIONAL (Default: no limit)
    min_memory = 512,

    # This field can be used to check the service's readiness after the service has started,
    # to confirm that it is ready to receive connections and traffic
    # OPTIONAL (Default: no ready conditions)
    ready_conditions = ReadyCondition(...),

    # This field is used to specify custom labels at the container level in Docker and Pod level in Kubernetes.
    # For Docker, the label syntax and format will follow: "com.kurtosistech.custom.key": "value"
    # For Kubernetes, the label syntax & format will follow: kurtosistech.com.custom/key=value

    # Labels must follow the label standards outlined in [RFC-1035](https://datatracker.ietf.org/doc/html/rfc1035), 
    # meaning that both the label key and label value must contain at most 63 characters, contain only lowercase 
    # alphanumeric characters, dashes (-), underscores (_) or dots (.), start with an alphabetic character, and end with an alphanumeric character.
    # Empty value and capital letters are valid on label values.
    # OPTIONAL
    labels = {
        "key": "value",
        # Examples
        "name": "alice",
    }
)
```
The `ports` dictionary argument accepts a key value pair, where `key` is a user defined unique port identifier and `value` is a [PortSpec][port-spec] object.

The `files` dictionary argument accepts a key value pair, where `key` is the path where the contents of the artifact will be mounted to and `value` is a [Directory][directory] object or files artifact name.
Using a `Directory` object with `artifact_name` is strictly equivalent to directly using the files artifact name as the value of the dictionary. This is just to simplify usage.

You can view more information on [configuring the `ReadyCondition` type here][ready-condition].

:::tip
If you are trying to use a more complex versions of `cmd` and are running into issues, we recommend using `cmd` in combination with `entrypoint`. You can
set the `entrypoint` to `["/bin/sh", "-c"]` and then set the `cmd` to the command as you would type it in your shell. For example, `cmd = ["echo foo | grep foo"]`
:::

:::tip Example of a rendered label
If you have defined the pod label key:value pair in `ServiceConfig` to be:

```py
config = ServiceConfig(
    ...
    labels = {
        "name": "alice",
	"age": "20",
	"height": "175"
    }
)
```

then the labels for the pods on Kubernetes will look like:
```
labels:
	kurtosistech.com.custom/name=alice
	kurtosistech.com.custom/age=20
	kurotsistech.com.custom/height=175
```

while on Docker, the container labels will look like:
```
labels:
	"com.kurtosistech.custom.name": "alice"
	"com.kurtosistech.custom.age": "20"
	"com.kurtosistech.custom.height": "175"
```
:::

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[add-service-reference]: ./plan.md#add_service
[directory]: ./directory.md
[port-spec]: ./port-spec.md
[ready-condition]: ./ready-condition.md
