---
title: service inspect
sidebar_label: service inspect
slug: /service-inspect
---

To view detailed information about a given service, including its status and attributes, run:

```bash
kurtosis service inspect $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.

Running the above command will print detailed information about:

- The service name and UUID
- The service status (running or stopped)
- The service container image name
- The service private ports and their public mapping
- The service container ENTRYPOINT, CMD and ENV

By default, the service UUID is shortened. To view the full UUID of your service, add the following flag:
* `--full-uuid`
