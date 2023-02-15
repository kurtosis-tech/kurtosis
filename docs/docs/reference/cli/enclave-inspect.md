---
title: Inspect an enclave
sidebar_label: Enclave inspect
slug: /enclave-inspect
---

To view detailed information about a given enclave, including its status and contents, run:

```bash
kurtosis enclave inspect [flags] $THE_ENCLAVE_IDENTIFIER 
```

This will print detailed information about:

- The enclave's status (running or stopped)
- The services inside the enclave (if any), and the information for accessing those services' ports from your local machine

By default, UUIDs are shortened. To view the full UUIDs of your resources, add the following flag:
* `--full-uuids`

Read more about resource identifiers in Kurtosis [here](../resource-identifier.md).