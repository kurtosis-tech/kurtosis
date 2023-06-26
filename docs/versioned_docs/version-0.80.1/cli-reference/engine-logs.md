---
title: engine logs
sidebar_label: engine logs
slug: /engine-logs
---

To get logs for all existing (stopped or running) engines, use:

```bash
kurtosis engine logs $OUTPUT_DIRECTORY
```

which will dump all the logs of the engine container to the directory specified by `$OUTPUT_DIRECTORY`. If a `$OUTPUT_DIRECTORY` is not specified, Kurtosis will default to writing the logs in a folder name following the schema `kurtosis-engine-logs--TIMESTAMP` in the working directory.
