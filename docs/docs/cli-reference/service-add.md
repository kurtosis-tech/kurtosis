---
title: service add
sidebar_label: service add
slug: /service-add
---

To add a service to an enclave, run:

```bash
kurtosis service add $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $CONTAINER_IMAGE
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.
Note, the service identifier needs to be formatted according to RFC 1035. Specifically, 1-63 lowercase alphanumeric characters with dashes and cannot start or end with dashes. Also service names
have to start with a lowercase alphabet.

Much like `docker run`, this command has multiple options available to customize the service that's started:

1. The `--cmd` flag can be used to override the default command that the container runs
1. The `--entrypoint` flag can be passed in to override the binary the service runs
1. The `--env` flag can be used to specify a set of environment variables that should be set when running the service
1. The `--ports` flag can be used to set the ports that the service will listen on
1. The `--privileged` flag allows Docker-only `privileged`, `bind_mounts`, `host_pid_namespace`, and `host_cgroup_namespace` fields when using `--json-service-config`. This is an allow flag: it does not make the service privileged by itself. The JSON service config must explicitly include `privileged: true`, `bind_mounts`, `host_pid_namespace: true`, or `host_cgroup_namespace: true`.

To override the service's CMD, add a `--` after the image name and then pass in your CMD args like so:

```bash
kurtosis service add --entrypoint sh my-enclave test-service alpine -- -c "echo 'Hello world'"
```

Alternatively, if you have an existing service config in JSON format (for example, one that was output using `kurtosis service inspect`), you can use the `--json-service-config` flag to add a service using that config:

```bash
kurtosis service add my-enclave test-service --json-service-config ./my-service-config.json
```

To read the JSON config from stdin, use:

```bash
kurtosis service add my-enclave test-service --json-service-config - < ./my-service-config.json
```

:::note Override
When using `--json-service-config`, the standard flags and args like `--image`, `--cmd`, `--entrypoint`, `--env`, and `$CONTAINER_IMAGE` will be ignored in favor of the provided config.
:::

:::note Privileged services
Services that use `privileged: true`, `bind_mounts`, `host_pid_namespace: true`, or `host_cgroup_namespace: true` require `--privileged` or `allow-privileged-mode: true` in the current cluster's `kurtosis-config.yml`. These fields are Docker-only and are rejected on Kubernetes.
:::
