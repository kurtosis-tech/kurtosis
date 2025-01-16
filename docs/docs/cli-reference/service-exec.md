---
title: service exec
sidebar_label: service exec
slug: /service-exec
---

To run a specific shell command inside a service container, run:

```bash
kurtosis service exec [--user $CONTAINER_USER] $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER 'COMMAND'
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.

Optionally pass `--user` flag to `exec` with $CONTAINER_USER, to execute the command on the container as that user. This only works for the docker case, the kubernetes case will fail if used. Omitting `--user` will default to `root`.

The specified command should be appropriately quoted and will be passed as it is to the shell interpreter of the running service container.

If the command returns a non-zero exit code, Kurtosis CLI will print an error and also return a non-zero exit code.
