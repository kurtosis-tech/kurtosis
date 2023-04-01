---
title: enclave add
sidebar_label: enclave add
slug: /enclave-add
---

Your distributed applications run in [enclaves][enclaves-reference]. They are isolated from each other, to ensure they don't interfere with each other. To create a new, empty enclave, simply run:

```bash
kurtosis enclave add
```

To create enclaves that support [subnetworks][subnetworks] use the `--with-subnetworks` flag.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[enclaves-reference]: ../concepts-reference/enclaves.md
[subnetworks]: ../concepts-reference/subnetworks.md
