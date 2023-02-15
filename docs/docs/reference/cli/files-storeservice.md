---
title: Store a file or folder in an enclave
sidebar_label: Files store service
slug: /files-storeservice
---

To instruct Kurtosis to copy a file or folder from a given absolute filepath in a given service and store it in the enclave for later use (e.g. with 'service add'), use:

```bash
kurtosis files storeservice [flags] $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $ABSOLUTE_SOURCE_FILEPATH
```

