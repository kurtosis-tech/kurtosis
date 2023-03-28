---
title: engine logs
sidebar_label: engine logs
slug: /engine-logs
---

To get logs for all existing engines, use:

```bash
kurtosis engine logs $OUTPUT_DIRECTORY
```

This will dump all the logs of the container to the directory specified by `$OUTPUT_DIRECTORY`. If it isn't `$OUTPUT_DIRECTORY` isn't specified
Kurtosis will default to writing the logs in a folder called `kurtosis-engine-logs`
