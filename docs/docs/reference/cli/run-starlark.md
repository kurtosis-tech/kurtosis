---
title: Run Starlark
sidebar_label: Run Starlark
slug: /run-starlark
sidebar_position: 5
---

### Run Starlark
A single Starlark script can be ran with:

```bash
kurtosis run script.star
```

Adding the `--dry-run` flag will print the changes without executing them.

A [Kurtosis package](../packages.md) on your local machine can be run with:

```bash
kurtosis run /path/to/package/on/your/machine
```

A [runnable Kurtosis package](../packages.md) published to GitHub can be run like so:

```bash
kurtosis run github.com/package-author/package-repo
```

Arguments can be provided to a Kurtosis package (either local or from GitHub) by passing a JSON-serialized object with args argument, which is the second positional argument you pass to `kurtosis run` like:

```bash
# Local package
kurtosis run /path/to/package/on/your/machine '{"company":"Kurtosis"}'

# GitHub package
kurtosis run github.com/package-author/package-repo '{"company":"Kurtosis"}'
```
