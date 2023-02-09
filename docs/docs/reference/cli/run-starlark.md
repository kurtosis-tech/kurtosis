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

:::tip
If you want to run a non-master branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

Arguments can be provided to a Kurtosis package (either local or from GitHub) by passing a JSON-serialized object with args argument, which is the second positional argument you pass to `kurtosis run` like:

```bash
# Local package
kurtosis run /path/to/package/on/your/machine '{"company":"Kurtosis"}'

# GitHub package
kurtosis run github.com/package-author/package-repo '{"company":"Kurtosis"}'
```

This command has options available to customize its execution:

1. The `--dry-run` flag can be used to print the changes proposed by the script without executing them
1. The `--parallelism` flag can be used to specify to what degree of parallelism certain commands can be run. For example: If the script contains `add_services` and is run with `--parallelism 100`, up to 100 services will be run at one time.