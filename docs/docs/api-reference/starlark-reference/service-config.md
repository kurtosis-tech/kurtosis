---
title: ServiceConfig
sidebar_label: ServiceConfig
---

The `ServiceConfig` object constructor configures the parameters of a container for the [`plan.add_service`][add-service-reference] function.


Signature
---------
```python
ServiceConfig(
    image,
    ports = {},
    env_vars = {},
    files = {},
    entrypoint = None,
    cmd = None,
    private_ip_address_placeholder = "KURTOSIS_IP_ADDRESS_PLACEHOLDER",
    min_cpu = None,
    max_cpu = None,
    min_memory = None,
    max_memory = None,
    ready_conditions = [],
    labels = {},
    user = None,
    tolerations = None,
    node_selectors = None,
)
```

| Property | Description |
| --- | --- |
| **image**<br/>_string\|[ImageSpec][image-spec]\|[ImageBuildSpec][image-build-spec]\|[NixBuildSpec][nix-build-spec]_ | The container image that will be used to back the service.<br/><br/>This can a simple container image string just like you'd feed to `docker run`, or a more complex specification instructing Kurtosis to build an image on demand.<br/><br/>If a `string` is provided, the string will be treated as an image identifier just like you'd use in `docker run`. `name = "IMAGE"` is merely syntactic sugar for `ImageSpec(name = "IMAGE")`.<br/><br/>If an [`ImageSpec`][image-spec] is provided, then the information provided in the `ImageSpec` will be used to pull the image.<br/><br/>If an [`ImageBuildSpec`][image-build-spec] is provided, Kurtosis will build the image using the provided Dockerfile information upon every run.<br/><br/>If a [`NixBuildSpec`][nix-build-spec] is provided, Kurtosis will build the image using the provided Nix Flake information  |
| **ports**<br/> _dict\[string, [PortSpec][port-spec]]_ | A mapping defining the ports that the container will be listening on, with the keys being user-friendly port names and the values being [`PortSpec`][port-spec] objects defining information about the ports.<br/><br/>The port names specified in this map will correspond to the ports on the [`Service`][service] object returned by [`plan.add_service`][add-service-reference].<br/><br/>Kurtosis will automatically check to ensure all declared TCP ports are open and ready for connections upon container startup. This behaviour can be disabled in the [`PortSpec`][port-spec] configuration.<br/><br/>If no ports are provided, no ports will be exposed unless the container image's Dockerfile has an `EXPOSE` directive.  |
| **env_vars**<br/> _dict\[string, string]_ | A mapping of environment variable names to values that should be set when running the container.<br/>Any [future reference strings][future-references] used in the value will be replaced at runtime with the actual value. |
| **files**<br/> _dict\[string, string\|[Directory][directory]]_ | A mapping of filepaths to the files that should be mounted there.<br/><br/>If a string value is provided, the [files artifact][files-artifacts] matching the same name will be mounted at that path. This is syntactic sugar for the more verbose `Directory(artifact_names=["artifact-name"])`.<br/><br/>If a [`Directory`][directory] value is provided, the provided configuration will be used to configure the mounted files. |
| **entrypoint**<br/>_list\[string]_ | Overrides the `ENTRYPOINT` setting of the container image.<br/><br/>If `None` is provided, then the default `ENTRYPOINT` setting on the container image will be used.<br/><br/>If `list[string]` is provided, then the `ENTRYPOINT` will be overridden with the given strings. Any [future reference strings][future-references] used will be replaced at runtime with the actual value. |
| **cmd**<br/>_list\[string]_ | Overrides the `CMD` setting of the container image.<br/><br/>If `None` is provided, then the default `CMD` setting on the container image will be used.<br/><br/>If `list[string]` is provided, then the `CMD` will be overridden with the given strings. Any [future reference strings][future-references] used will be replaced at runtime with the actual value.<br/><br/>**💁 TIP:** If you are trying to use a more complex versions of `cmd` and are running into issues, we recommend using `cmd` in combination with `entrypoint`. You can set `entrypoint = ["/bin/sh", "-c"]` and then set the `cmd` to the command as you would type it in your shell (e.g. `cmd = ["echo foo | grep foo"]`). |
| **private_ip_address_placeholder**<br/>_string_ | On occasion, the `entrypoint`, `cmd`, and `env_vars` properties may need to know the IP address of the container before it exists. If the value of `private_ip_address_placeholder` is included in these properties, the placeholder value will be replaced with the container's actual IP address at runtime. |
| **min_cpu**<br/>_int_ | Controls the minimum amount of CPU that a service is allocated, in millicpus/millicores.<br/><br/>If set to `None`, then the service will not have any CPU set aside just for it.<br/><br/>If set to a non-`None` integer, the service will be allocated the specified number of millicpus.<br/><br/>**⚠️ CAUTION:** Docker doesn't have a setting to enforce this, so this option will only be used on Kubernetes. |
| **max_cpu**<br/>_int_ | Controls the maximum amount of CPU that a service is allowed to consume, in millicpus/millicores.<br/><br/>If set to `None`, then the service will not have any CPU limits set for it.<br/><br/>If set to a non-`None` integer, the service will be capped to the specified number of millicpus. |
| **min_memory**<br/>_int_ | Controls the minimum amount of memory that a service is allocated, in megabytes.<br/><br/>If set to `None`, then the service will not have any memory set aside just for it.<br/><br/>If set to a non-`None` integer, the service will be allocated the specified number of megabytes.<br/><br/>**⚠️ CAUTION:** Docker doesn't have a setting to enforce this, so this option will only be used on Kubernetes. |
| **max_memory**<br/>_int_ | Controls the maximum amount of memory that a service is allowed to consume, in megabytes.<br/><br/>If set to `None`, then the service will not have any memory limits set for it.<br/><br/>If set to a non-`None` integer, the service will be capped to the specified number of megabytes. |
| **ready_conditions**<br/>_list[[ReadyCondition][ready-condition]]_ | Defines a set of [`ReadyCondition`][ready-condition] checks that must pass before Kurtosis considers the service started.<br/><br/>If any ReadyCondition doesn't pass, Kurtosis considers the service as having failed to start. |
| **labels**<br/>_dict\[string, string]_ | A mapping of string keys to values used to attach custom labels to the containers that Kurtosis starts. This can be useful if you have automation consuming the Kurtosis enclave.<br/><br/>For Docker, providing `labels={"KEY":"VALUE"}` will result in the container label `"com.kurtosistech.custom.KEY": "VALUE"`.<br/><br/>For Kubernetes, providing `labels={"KEY":"VALUE"}` will result in the Pod label `kurtosistech.com.custom/KEY=VALUE`. |
| **user**<br/>_[User][user]_ | The user to start the container with. This is useful when you need to override the user ID or group ID that a container runs under. |
| **tolerations**<br/>_list[[Toleration][toleration]]_ | When running on Kubernetes, creates Kubernetes [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for each entry in the given list. Requires at least one untainted node to function.<br/><br/>**⚠️ WARNING:** This is an experimental feature that might get replaced with a better abstraction in the future.<br/><br/>**📝 NOTE:** Has no effect on Docker. |
| **node_selectors**<br/>_dict\[string, string]_ | Refers to Kubernetes [Node Selectors](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/#create-a-pod-that-gets-scheduled-to-your-chosen-node), used to control which nodes a Pod is scheduled on.<br/>For example, `node_selectors = {"disktype": "ssd"}` could be used to schedule the service only on nodes with the `disktype=ssd` label.<br/><br/>**⚠️ WARNING:** This is an experimental feature that might get replaced with a better abstraction in the future.<br/><br/>**📝 NOTE:** Has no effect on Docker. |

Basic Examples
-----------
You'll most frequently use the `image`, `ports`, `env_vars`, and `files` args. Occasionally, you'll need the `entrypoint` and `cmd` args.

### Postgres with preseeded data

```python
init_sql = plan.render_templates({
    "init.sql": struct(
        template = """
CREATE TABLE students (
id SERIAL PRIMARY KEY,
name VARCHAR(100),
age INTEGER
);

INSERT INTO students (name, age) VALUES
('Mkyong', 40),
('Ali', 28),
('Teoh', 18);
""",
        data={}
    ),
})

plan.add_service(
    name = "postgres",
    config = ServiceConfig(
        image = "postgres",
        ports = {
            "postgres": PortSpec(number = 5432, application_protocol = "postgresql"),
        },
        env_vars = {
            "POSTGRES_USER": "teacher",
            "POSTGRES_PASSWORD": "bookworm",
            "POSTGRES_DB": "students",
        },
        files = {
            "/docker-entrypoint-initdb.d": init_sql,
        },
    )
)
```

### Persisting data through service restarts
The `Directory.persistent_key` property can be used to indicate that data should be persisted through service restarts:

```python
plan.add_service(
    name = "postgres",
    config = ServiceConfig(
        image = "postgres",
        ports = {
            "postgres": PortSpec(number = 5432, application_protocol = "postgresql"),
        },
        env_vars = {
            "POSTGRES_USER": "teacher",
            "POSTGRES_PASSWORD": "bookworm",
            "POSTGRES_DB": "students",
        },
        files = {
            "/var/lib/postgresql/data": Directory(persistent_key = "pgdata")
        },
    )
)
```

<!-- 
TODO TODO TODO
### Overriding ENTRYPOINT and CMD
Sometimes, you'll need to override the default `ENTRYPOINT` and `CMD` of a container image. This can be done using the `entrypoint` and `cmd` args like so:

```python

```
-->

Advanced Examples
--------------
### Assigning custom labels
Suppose you've defined the following `ServiceConfig`:

```python
ServiceConfig(
    image = "postgres",
    labels = {
        "name": "alice",
        "age": "20",
        "height": "175",
    }
)
```

When running on Kubernetes, the labels for the Pod running the service will look like:
```
kurtosistech.com.custom/name=alice
kurtosistech.com.custom/age=20
kurotsistech.com.custom/height=175
```

When running on Docker, the container labels will look like:
```
"com.kurtosistech.custom.name": "alice"
"com.kurtosistech.custom.age": "20"
"com.kurtosistech.custom.height": "175"
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[add-service-reference]: ./plan.md#add_service
[directory]: ./directory.md
[port-spec]: ./port-spec.md
[ready-condition]: ./ready-condition.md
[locators]: ../../advanced-concepts/locators.md
[package]: ../../advanced-concepts/packages.md
[user]: ./user.md
[toleration]: ./toleration.md
[nix-support]: ./nix-support.md
[future-references]: ../../advanced-concepts/future-references.md
[files-artifacts]: ../../advanced-concepts/files-artifacts.md
[service]: ./service.md
[image-spec]: ./image-spec.md
[image-build-spec]: ./image-build-spec.md
[nix-build-spec]: ./nix-support.md
