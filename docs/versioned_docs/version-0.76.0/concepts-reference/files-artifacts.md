---
title: Files Artifacts
sidebar_label: Files Artifacts
---

Kurtosis enclaves can store files for later use. Files stored in a Kurtosis enclave are stored as compressed TGZ files. These TGZs are called "files artifacts".

For example, a user can upload files on their machine to an enclave like so:

```bash
kurtosis files upload $SOME_PATH
```

:::info
If `$SOME_PATH` is a file, that single file will be packaged inside the files artifact. If `$SOME_PATH` is a directory, all of the directory's contents will be packaged inside the files artifact.
:::

Doing so will return a randomly-generated ID and name that can be used to reference the files artifact for later use.

For example, the `--files` flag of `kurtosis service add` can be used to mount the contents of a files artifact at specified location. This command will mount the contents of files artifact `test-artifact` at the `/data` directory:

```bash
kurtosis service add "some-enclave" "some-service-name" --files "/data:test-artifact"
```

The same files artifact can be reused many times because the contents of a files artifact is copied when it is used.