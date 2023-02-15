---
title: Starting the engine
sidebar_label: Engine start
slug: /engine-start
---

THe Kurtosis engine starts automatically when you run [`kurtosis enclave add`](./enclave-add.md), but there may be times where the engine has been stopped and you may need to start it again (e.g. starting the engine on a specific version). To do so, run:

```bash
kurtosis engine start [flags]
```
This command will do nothing if the Kurtosis engine is already running.

You may optionally pass in the following flags with this command:
* `--log-level`: The level that the started engine should log at. Options include: panic, fatal,error, warning, info, debug, or trace. The engine logs at the "info" level by default.
* `--version`: The version (Docker tag) of the Kurtosis engine that should be started. If not set, the engine run start up with the default version.