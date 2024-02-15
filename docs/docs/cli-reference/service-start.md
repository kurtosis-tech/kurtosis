---
title: service start
sidebar_label: service start
slug: /service-start
---

Temporarily stopped services in an enclave can be restarted like so:

```bash
kurtosis service start $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.
