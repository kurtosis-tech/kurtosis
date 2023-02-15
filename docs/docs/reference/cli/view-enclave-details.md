---
title: View enclave details
sidebar_label: View enclave details
slug: /view-enclave-details
sidebar_position: 8
---

### View enclave details
To view detailed information about a given enclave, run:

```bash
kurtosis enclave inspect $THE_ENCLAVE_IDENTIFIER
```

This will print detailed information about:

- The enclave's status (running or stopped)
- The services inside the enclave (if any), and the information for accessing those services' ports from your local machine