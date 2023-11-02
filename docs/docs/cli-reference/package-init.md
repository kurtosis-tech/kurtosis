# kurtosis package init
This command initializes the current directory to be a [Kurtosis package][package] by creating a [`kurtosis.yml` file][kurtosis-yml] with the given package name.

```
Usage:
  kurtosis package init [flags] package_name
```

The `package_name` argument is the [locator][locators]to the package, in the format `github.com/USER/REPO`.

[package]: ../advanced-concepts/packages.md
[kurtosis-yml]: ../advanced-concepts/kurtosis-yml.md
[locators]: ../advanced-concepts/locators.md
[executable-package]: ../advanced-concepts/packages.md#runnable-packages