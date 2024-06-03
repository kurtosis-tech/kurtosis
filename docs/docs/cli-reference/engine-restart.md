---
title: engine restart
sidebar_label: engine restart
slug: /engine-restart
---

The CLI interacts with the Kurtosis engine, which is a very lightweight container. The CLI will start the engine container automatically for you and you should never need to start it manually, but you might need to restart the engine after a CLI upgrade. To do so, run:

```bash
kurtosis engine restart
```

You may optionally pass in the following flags with this command:
* `--log-level`: The level that the started engine should log at. Options include: `panic`, `fatal`, `error`, `warning`, `info`, `debug`, or `trace`. The engine logs at the `info` level by default.
* `--version`: The version (Docker tag) of the Kurtosis engine that should be started. If not set, the engine will start up with the default version.
* `--author`:  "The author (Docker username) of the Kurtosis engine that should be started (blank will start the kurtosistech version). The same author and version are used for the API Container & Files Artifact Expander.
* `--enclave-pool-size`: The size of the Kurtosis engine enclave pool. The enclave pool is a component of the Kurtosis engine that allows us to create and maintain 'n' number of idle enclaves for future use. This functionality allows to improve the performance for each new creation enclave request.

CAUTION: The `--enclave-pool-size` flag is only available for Kubernetes. The enclave pool is disabled for custom engine author for now.