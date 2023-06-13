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
