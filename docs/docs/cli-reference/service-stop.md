---
title: service stop
sidebar_label: service stop
slug: /service-stop
---

Services can be temporarily stopped in an enclave like so:

```bash
kurtosis service stop $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively.
