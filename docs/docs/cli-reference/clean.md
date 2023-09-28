---
title: clean
sidebar_label: clean
slug: /clean
---

The `clean` command is used to remove enclaves and related processes, often to free up resources.  

1. Removes stopped enclaves and stopped engine containers (if the -a flag is set, all running enclaves and engine containers will also be removed)
2. Removes all services within enclaves
3. Removes all unused kurtosis images (engine + logs aggregator + enclaves)
4. Removes all files artifacts and all docker volumes within those enclaves
```
  kurtosis clean [flags]
```
Flags:
1. The `a, --all` removes running enclaves as well
2. The `-h, --help` flag shows help for clean


NOTE: This will not stop the Kurtosis engine itself! To do so, use the [engine stop](./engine-stop.md) command.
