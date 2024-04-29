---
title: clean
sidebar_label: clean
slug: /clean
---

The `clean` command serves the purpose of freeing up resources from your local machine by removing unused Kurtosis images and any stopped enclaves (along with their contents). 

- Removes stopped enclaves and stopped engine containers
- Removes all services within stopped enclaves
- Removes all unused Kurtosis images (engine + logs aggregator + enclaves)
- Removes all files artifacts and all Docker volumes within stopped enclaves

```
  kurtosis clean [flags]
```
Flags:
1. The `-a, --all` removes running enclaves as well
2. The `-h, --help` flag shows help for clean


NOTE: This will not stop the Kurtosis engine itself! To do so, use the [engine stop](./engine-stop.md) command.
