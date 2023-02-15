---
title: View a service's logs
sidebar_label: View a service's logs
slug: /view-service-logs
sidebar_position: 12
---

### View a service's logs
To print the logs for a service, run:

```bash
kurtosis service logs $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

The service identifier(name or uuid) is printed upon inspecting an enclave.

The following optional arguments can be used:
1. `-f`, `-follow` can be added to continue following the logs, similar to `tail -f`.
1. `--match=text` can be used for filtering the log lines containing the text.
1. `--regex-match="regex"` can be used for filtering the log lines containing the regex. This filter will also work for text but will have degraded performance.
1. `-v`, `--invert-match` can be used to invert the filter condition specified by either `--match` or `--regex-match`. Log lines NOT containing the match will be returned.

Important: `--match` and `--regex-match` flags cannot be used at the same time. You should either use one or the other.