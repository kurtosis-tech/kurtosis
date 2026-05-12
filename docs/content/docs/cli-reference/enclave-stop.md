---
title: enclave stop
url: /enclave-stop/
linkTitle: enclave stop
---

To stop a particular enclave, use:

```bash
kurtosis enclave stop $THE_ENCLAVE_IDENTIFIER 
```
where `$THE_ENCLAVE_IDENTIFIER` is the enclave [identifier](../advanced-concepts/resource-identifier.md).

{{< callout type="warning" >}}
Enclaves that have been stopped cannot currently be restarted. The Kurtosis team is actively working on enabling this functionality.
{{< /callout >}}
