---
title: service delete
sidebar_label: service delete
slug: /service-delete
---

Services can be deleted from an enclave like so:

```bash
kurtosis service rm $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and service, respectively. 

**NOTE:** To avoid destroying debugging information, Kurtosis will leave removed services inside the Docker engine. They will be stopped and won't show up in the list of active services in the enclave, but you'll still be able to access them (e.g. using `service logs`) by their service GUID (available via `enclave inspect`).