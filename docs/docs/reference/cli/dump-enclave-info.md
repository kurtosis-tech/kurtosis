---
title: Dump enclave information to disk
sidebar_label: Dump enclave information to disk
slug: /dump-enclave-info
sidebar_position: 9
---

### Dump enclave information to disk
You'll likely need to store enclave logs to disk at some point. You may want to have a log package if your CI fails, or you want to send debugging information to [the author of a Kurtosis package][packages-reference]. Whatever the case may be, you can run:

```bash
kurtosis enclave dump $THE_ENCLAVE_IDENTIFIER $OUTPUT_DIRECTORY
```

You'll get the container logs & configuration in the output directory for further analysis & sharing.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[packages-reference]: ../packages.md