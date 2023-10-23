---
title: package init
sidebar_label: package init
slug: /package-init
---

The `package init` command converts the current directory into a [Kurtosis package][package] by generating a new [`kurtosis.yml`][kurtosis-yml] file using the given package name.

```
kurtosis package init $PACKAGE_NAME
```

The optional `$PACKAGE_NAME` argument is the [locator][locators] to the package, in the format `github.com/USER/REPO`. If an argument is not passed, the package locator will simply be: `github.com/example-org/example-package` by default.

<!---------------------->
[package]: ../concepts-reference/packages.md
[kurtosis-yml]: ../concepts-reference/kurtosis-yml.md
[locators]: ../concepts-reference/locators.md
[executable-package]: ../concepts-reference/packages.md#runnable-packages
