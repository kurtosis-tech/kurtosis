---
title: enclave inspect
sidebar_label: enclave inspect
slug: /enclave-inspect
---

To view detailed information about a given enclave, including its status and contents, run:

```bash
kurtosis enclave inspect $THE_ENCLAVE_IDENTIFIER 
```

where `$THE_ENCLAVE_IDENTIFIER` is the [resource identifier](../concepts-reference/resource-identifier.md) for the enclave.

Running the above command will print detailed information about:

- The enclave's status (running or stopped)
- The services inside the enclave (if any), their status, and the information for accessing those services' ports from your local machine
- Any files artifacts registered within the specified enclave

By default, UUIDs are shortened. To view the full UUIDs of your resources, add the following flag:
* `--full-uuids`

