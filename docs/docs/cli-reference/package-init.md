# kurtosis package init
This command initializes the current directory to be a [Kurtosis package][package] by creating a [`kurtosis.yml` file][kurtosis-yml] with the given package name.

```
Usage:
  kurtosis package init [flags] package_name
```

The `package_name` argument is the [locator][locators]to the package, in the format `github.com/USER/REPO`.

[package]: ../concepts-reference/packages.md
[kurtosis-yml]: ../concepts-reference/kurtosis-yml.md
[locators]: ../concepts-reference/locators.md
[executable-package]: ../concepts-reference/packages.md#runnable-packages