---
title: clean
sidebar_label: clean
slug: /clean
---

The `clean` command is used to remove enclaves and related processes, often to free up resources. By default this command affects stopped enclaves. If the -a flag is set, all running enclaves and engine containers will be removed.  

- Removes stopped enclaves and stopped engine containers
- Removes all services within enclaves
- Removes all unused Kurtosis images (engine + logs aggregator + enclaves)
- Removes all files artifacts and all docker volumes within those enclaves
```
  kurtosis clean [flags]
```
Flags:
1. The `a, --all` removes running enclaves as well
2. The `-h, --help` flag shows help for clean


NOTE: This will not stop the Kurtosis engine itself! To do so, use the [engine stop](./engine-stop.md) command.
