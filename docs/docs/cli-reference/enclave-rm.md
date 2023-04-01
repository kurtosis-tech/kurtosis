---
title: enclave rm
sidebar_label: enclave rm
slug: /enclave-rm
---

To remove an enclave and all resources associated with that particular enclave, use:

```bash
kurtosis enclave rm $THE_ENCLAVE_IDENTIFIER 
```
where `$THE_ENCLAVE_IDENTIFIER` is the enclave [identifier](../concepts-reference/resource-identifier.md).

Note that this command will only remove stopped enclaves. To destroy a running enclave, pass the `-f`/`--force` flag.