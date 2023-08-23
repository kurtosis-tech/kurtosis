---
title: clean
sidebar_label: clean
slug: /clean
---

```console
Removes stopped enclaves (and live ones if the 'all' flag is set), as well as stopped engine containers

Usage:
  kurtosis clean [flags]

Flags:
  -a, --all    If set, removes running enclaves as well
  -h, --help   help for clean
```

NOTE: This will not stop the Kurtosis engine itself! To do so, use the [engine stop](./engine-stop.md) command.
