---
title: package init
sidebar_label: package init
slug: /package-init
---

The `package init` command converts the current directory into a [Kurtosis package][package] by generating a new [`kurtosis.yml`][kurtosis-yml] file and a `main.star` file using the given package name, if they do not already exist.

```
kurtosis package init $PACKAGE_NAME
```

The optional `$PACKAGE_NAME` argument is the [locator][locators] to the package, in the format `github.com/USER/REPO`. If an argument is not passed, the package locator will simply be: `github.com/example-org/example-package` by default.

[package]: ../advanced-concepts/packages.md
[kurtosis-yml]: ../advanced-concepts/kurtosis-yml.md
[locators]: ../advanced-concepts/locators.md
[executable-package]: ../advanced-concepts/packages.md#runnable-packages
