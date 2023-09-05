---
title: enclave dump
sidebar_label: enclave dump
slug: /enclave-dump
---

You will likely need to store enclave logs to disk at some point. You may want to have a log package if your CI fails, or you want to send debugging information to [the author of a Kurtosis package][packages-reference]. Whatever the case may be, you can run:

```bash
kurtosis enclave dump $THE_ENCLAVE_IDENTIFIER $OUTPUT_DIRECTORY
```
where the `$THE_ENCLAVE_IDENTIFIER` is the [resource identifier](../concepts-reference/resource-identifier.md) for an enclave.

You will get the container logs & configuration in the output directory for further analysis & sharing.

If you don't specify the `$OUTPUT_DIRECTORY` Kurtosis will dump it to a directory with a name following the `ENCLAVE_NAME--ENCLAVE_UUID` scheme in the
current working directory.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[packages-reference]: ../concepts-reference/packages.md
