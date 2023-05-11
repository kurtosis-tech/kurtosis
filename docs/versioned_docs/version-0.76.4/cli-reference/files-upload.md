---
title: files upload
sidebar_label: files upload
slug: /files-upload
---

Files can be stored as a [files artifact][files-artifacts] inside an enclave by uploading them:

```bash
kurtosis files upload $THE_ENCLAVE_IDENTIFIER $PATH_TO_FILES
```

where `$THE_ENCLAVE_IDENTIFIER` is the [resource identifier][resource-identifier] for the enclave.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[files-artifacts]: ../concepts-reference/files-artifacts.md
[resource-identifier]: ../concepts-reference/resource-identifier.md
