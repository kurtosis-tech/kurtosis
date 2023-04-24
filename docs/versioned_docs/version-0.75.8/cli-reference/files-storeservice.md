---
title: files storeservice
sidebar_label: files storeservice
slug: /files-storeservice
---

To instruct Kurtosis to copy a file or folder from a given absolute filepath in a given service and store it in the enclave for later use (e.g. with [`service add`](./service-add.md)), use:

```bash
kurtosis files storeservice $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $ABSOLUTE_SOURCE_FILEPATH
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and service, respectively. 