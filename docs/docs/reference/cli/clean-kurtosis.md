---
title: Clean Kurtosis
sidebar_label: Clean Kurtosis
slug: /clean-kurtosis
sidebar_position: 16
---

### Clean Kurtosis
Kurtosis defaults to leaving enclave artifacts (containers, volumes, etc.) around so that you can refer back them for debugging. To clean up artifacts from stopped enclaves, run:

```bash
kurtosis clean
```

To remove artifacts from _all_ enclaves (including running ones), add the `-a`/`--all` flag.

NOTE: This will not stop the Kurtosis engine itself! To do so, see "Stopping the engine" above.