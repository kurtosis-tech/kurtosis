---
title: enclave add
sidebar_label: enclave add
slug: /enclave-add
---

Your distributed applications run in [enclaves][enclaves-reference]. They are isolated from each other, to ensure they don't interfere with each other. To create a new, empty enclave, simply run:

```bash
kurtosis enclave add
```

1. The `--production` flag can be used to make sure services restart in case of failure (default behavior is not restart)

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[enclaves-reference]: ../advanced-concepts/enclaves.md
