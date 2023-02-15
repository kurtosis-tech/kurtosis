---
title: Remove an enclave
sidebar_label: Enclave remove
slug: /enclave-rm
---

To remove an enclave and all resources associated with that particular enclave, use:

```bash
kurtosis enclave rm $THE_ENCLAVE_IDENTIFIER 
```

Note that this command will only remove stopped enclaves. To destroy a running enclave, pass the `-f`/`--force` flag.