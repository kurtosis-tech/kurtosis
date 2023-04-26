---
title: files storeweb
sidebar_label: files storeweb
slug: /files-storeweb
---

To download an archive file from the given URL and store it in the enclave for later use (e.g. with [`service add`](./service-add.md)), use:

```bash
kurtosis files storeweb $THE_ENCLAVE_IDENTIFIER $URL
```

where `$THE_ENCLAVE_IDENTIFIER` is the [resource identifier](../concepts-reference/resource-identifier.md) for the enclave.