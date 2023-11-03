---
title: port print
sidebar_label: port print
slug: /port-print
---

To print information about the PortSpec for Service, run:

```bash
kurtosis port print $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $PORT_ID
```
where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../advanced-concepts/resource-identifier.md) for the enclave and service, respectively. The `$PORT_ID` is the unique port identifier assigned to the port using [`ServiceConfig`](../api-reference/starlark-reference/service-config.md) on starlark.