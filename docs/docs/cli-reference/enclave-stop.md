---
title: enclave stop
sidebar_label: enclave stop
slug: /enclave-stop
---

To stop a particular enclave, use:

```bash
kurtosis enclave stop $THE_ENCLAVE_IDENTIFIER 
```
where `$THE_ENCLAVE_IDENTIFIER` is the enclave [identifier](../concepts-reference/resource-identifier.md).

:::caution
Enclaves that have been stopped cannot currently be restarted. The Kurtosis team is actively working on enabling this functionality.
:::
