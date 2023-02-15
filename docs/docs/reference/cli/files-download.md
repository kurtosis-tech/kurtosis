---
title: Download files from an enclave
sidebar_label: Files download
slug: /files-download
---

To download a files artifact using a resource identifier identifier (e.g. name, uuid, shortened uuid) from an enclave to the host machine, use:

```bash
kurtosis files download  [flags] $THE_ENCLAVE_IDENTIFIER $THE_ARTIFACT_IDENTIFIER $FILE_DESTINATION_PATH
```

The file downloaded will not be extracted by default. If you would prefer the file be extracted upon download, pass in the `--no-extract` flag.