---
title: engine start
sidebar_label: engine start
slug: /engine-start
---

The Kurtosis engine starts automatically when you run any command that interacts with the engine, such as [`kurtosis enclave add`](./enclave-add.md), but there may be times where the engine has been stopped and you may need to start it again (e.g. starting the engine on a specific version). To do so, run:

```bash
kurtosis engine start
```
This command will do nothing if the Kurtosis engine is already running.

You may optionally pass in the following flags with this command:
* `--log-level`: The level that the started engine should log at. Options include: `panic`, `fatal`, `error`, `warning`, `info`, `debug`, or `trace`. The engine logs at the `info` level by default.
* `--version`: The version (Docker tag) of the Kurtosis engine that should be started. If not set, the engine will start up with the default version.
* `--enclave-pool-size`: The size of the Kurtosis engine enclave pool. The enclave pool is a component of the Kurtosis engine that allows us to create and maintain 'n' number of idle enclaves for future use. This functionality allows to improve the performance for each new creation enclave request.
* `--github-auth-token`: The auth token to use for authorizing GitHub operations. If set, this will override the currently logged in GitHub user from `kurtosis github login`, if one exists. Note, this token does not persist when restarting the engine.
* `--log-retention-period`: The duration in which Kurtosis engine will keep logs for. The engine will remove any logs beyond this period. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (eg. "300ms", "-1.5h" or "2h45m", "168h") The default is set to 1 week (168h). NOTE: Currently, Kurtosis only supports setting retention on weekly intervals. Ongoing work is occurring to make this interval more granular - see https://github.com/kurtosis-tech/kurtosis/pull/2534

CAUTION: The `--enclave-pool-size` flag is only available for Kubernetes.