---
title: Delete an enclave
sidebar_label: Delete an enclave
slug: /delete-enclave
sidebar_position: 10
---

### Delete an enclave
To delete an enclave and everything inside of it, run:

```bash
kurtosis enclave rm $THE_ENCLAVE_IDENTIFIER
```

Note that this will only delete stopped enclaves. To delete a running enclave, pass the `-f`/`--force` flag.