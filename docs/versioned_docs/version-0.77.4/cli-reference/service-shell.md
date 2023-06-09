---
title: service shell
sidebar_label: service shell
slug: /service-shell
---

To get access to a shell on a given service container, run:

```bash
kurtosis service shell $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and service, respectively.

Adding the `--exec` flag will run the specified command instead of bash/sh on the container; use it as follows -

```bash
kurtosis service shell $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER --exec 'command to run'
```