---
title: Delete a service from an enclave
sidebar_label: Delete a service from an enclave
slug: /delete-service
sidebar_position: 14
---

### Delete a service from an enclave
Services can be deleted from an enclave like so:

```bash
kurtosis service rm $THE_ENCLAVE_ID $THE_SERVICE_IDENTIFIER
```

**NOTE:** To avoid destroying debugging information, Kurtosis will leave removed services inside the Docker engine. They will be stopped and won't show up in the list of active services in the enclave, but you'll still be able to access them (e.g. using `service logs`) by their service GUID (available via `enclave inspect`).