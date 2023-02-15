---
title: Dumping information from an enclave
sidebar_label: Enclave dump
slug: /enclave-dump
---

You'll likely need to store enclave logs to disk at some point. You may want to have a log package if your CI fails, or you want to send debugging information to [the author of a Kurtosis package][packages-reference]. Whatever the case may be, you can run:

```bash
kurtosis enclave dump $THE_ENCLAVE_IDENTIFIER $OUTPUT_DIRECTORY
```

You'll get the container logs & configuration in the output directory for further analysis & sharing.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[packages-reference]: ../packages.md