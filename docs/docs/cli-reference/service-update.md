---
title: service update
sidebar_label: service update
slug: /service-update
---

To update an existing service in an enclave, run:

```bash
kurtosis service update $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER [flags]
```

where `$THE_ENCLAVE_IDENTIFIER` and `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.

This command updates a service in-place by modifying its configuration. Only the specified parameters will be changed — the rest of the service config will remain as-is.

Much like `docker run`, this command has multiple options available to customize the updated service:

1. The `--image` flag can be used to update the service’s container image
1. The `--entrypoint` flag can override the binary the service runs
1. The `--env` flag can be used to set or override environment variables. Env var overrides with the same key will override existing env vars.
1. The `--ports` flag can be used to add or override private port definitions. Port overrides with the same port id will override existing port bindings.
1. The `--files` flag can be used to mount new file artifacts. Files artifacts overrides with the same key will override existing files artifact mounts.
1. The `--cmd` flag can be used to override the CMD that is run when the container starts
1. The `--privileged` flag allows Docker-only `privileged`, `bind_mounts`, and `host_pid_namespace` fields to be preserved or enabled during the update. This is an allow flag: it does not make the service privileged by itself.

Example:

```bash
kurtosis service update my-enclave test-service \
  --image my-custom-image \
  --entrypoint my-binary \
  --env "FOO:bar,BAR:baz" \
  --ports "port1:8080/tcp"
```

:::note Restarted Container
This command replaces the existing service with a new container using the updated configuration. The service will be briefly stopped and restarted as part of this process.
:::

:::note Port wait
When you update a service, any custom `wait` configuration set on its ports will be cleared. All updated ports will have `wait=None` after this operation, regardless of their previous setting.
:::

:::note Privileged services
Updating a service that already has `privileged: true`, `bind_mounts`, or `host_pid_namespace: true` requires `--privileged` or `allow-privileged-mode: true` in the current cluster's `kurtosis-config.yml`, otherwise the update is rejected while the updated service config is interpreted. `service update` currently preserves and can enable these fields, but it cannot clear `privileged: true`, remove existing bind mounts, or unset `host_pid_namespace: true`. To remove privileged access today, recreate the service without those fields.
:::
