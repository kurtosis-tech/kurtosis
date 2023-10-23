---
title: package init
sidebar_label: package init
slug: /package-init
---

The `package init` command converts the current directory into a [Kurtosis package][package] by generating a new [`kurtosis.yml`][kurtosis-yml] file using the given package name.

```
Usage:
  kurtosis package init [flags] $PACKAGE_NAME
```

The mandatory `#PACKAGE_NAME` argument is the [locator][locators] to the package, in the format `github.com/USER/REPO`.

This command accepts the following flags:
- `--main`: indicates that the created package is an [executable package][executable-package], and generates a `main.star` if one does not already exist. If a `main.star` already exists, does nothing.

[package]: ../concepts-reference/packages.md
[kurtosis-yml]: ../concepts-reference/kurtosis-yml.md
[locators]: ../concepts-reference/locators.md
[executable-package]: ../concepts-reference/packages.md#runnable-packages
