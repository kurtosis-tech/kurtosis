---
title: files download
sidebar_label: files download
slug: /files-download
---

To download a [files artifact](../concepts-reference/files-artifacts.md) using a resource identifier (e.g. name, UUID, shortened UUID) from an enclave to the host machine, use:

```bash
kurtosis files download $THE_ENCLAVE_IDENTIFIER $THE_ARTIFACT_IDENTIFIER $FILE_DESTINATION_PATH
```
where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_ARTIFACT_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and file artifact, respectively. 

:::tip
The file downloaded will be extracted by default. If you would prefer the file not to be extracted upon download, pass in the `--no-extract` flag.
:::