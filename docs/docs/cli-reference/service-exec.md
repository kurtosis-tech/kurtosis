---
title: service exec
sidebar_label: service exec
slug: /service-exec
---

To run a specific shell command inside a service container, run:

```bash
kurtosis service exec $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER 'COMMAND'
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and service, respectively.

The specified command should be appropriately quoted and will be passed as it is to the shell interpreter of the running service container.

If the command returns a non-zero exit code, Kurtosis CLI will print an error and also return a non-zero exit code.

:::tip Special note about `service exec` vs `service shell --exec`
The main difference between `kurtosis service exec` and `kurtosis service shell --exec` is that the former expects a zero (0) exit code for the command and will otherwise throw an exception, whereas the latter returns gracefully whatever the exit code of the command was.

In short, `kurtosis service shell --exec` is designed for interactive sessions and will always return gracefully as long as the session terminated is ended successfully. Meanwhile, `kurtosis service exec` provides a way to run a single command and returns immediately, taking the status of the command into account.
:::
