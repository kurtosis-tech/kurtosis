---
title: Locators
sidebar_label: Locators
---

:::info
Locators are a part of the [Kurtosis packaging system][packages]. To read about the package system in detail, [see here][how-do-kurtosis-imports-work-explanation].
:::

A locator is how resources are identified when being imported in a Starlark script - namely by [`import_module`](../api-reference/starlark-reference/import-module.md) and [`read_file`](../api-reference/starlark-reference/read-file.md). There are two types of locators: absolute and relative.

### Absolute Locators
Absolute locators unambiguously identify a resource using a URL-like syntax. For example, this locator:

```
github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml/some-file.star
```

references a file inside a GitHub repo called `package-repo`, owned by `package-author`, that lives at the path `/path/to/directory-with-kurtosis.yml/some-file.star` relative to the root of the repo.

:::caution
A GitHub URL is **not** a valid locator, because GitHub adds extra `/blob/main` paths to the URL that don't reflect the file's path in the repo. For example, a GitHub URL of:

```
https://github.com/kurtosis-tech/kurtosis/blob/main/starlark/test.star
```

would be the following as a Kurtosis locator (dropping the `https://` and `/blob/main` part):

```
github.com/kurtosis-tech/kurtosis/starlark/test.star
```
:::

:::info
Locators can point to public or private GitHub repositories. Read the [Running Private Packages][running-private-packages]guide to learn how to enable private locators.
:::

### Important Package Restriction
If your Starlark script relies on local resources, such as files or packages available on your filesystem, then those resources *must* be part of a [Kurtosis package][packages]. 

For example, suppose we had a [Kurtosis package][packages] like so:

```
/
    package-repo/
        my-package/
            kurtosis.yml
            main.star
            helpers/
                random-script.star
        not-a-package/
            random-script.star
```

with a `kurtosis.yml` file like so:

```yaml
name: github.com/package-author/package-repo/my-package
```

In your `main.star` file, you would be able to import the `random-script.star` from the `helpers` subdirectory of `my-package` like so:

```python
# Valid
helpers = import_module("github.com/package-author/package-repo/my-package/helpers/random-script.star")
```

However, if you try to import `package-repo/not-a-package/random-script.star`, then it will not work because `package-repo/not-a-package/random-script.star` is not part of a package. In essence, the import statement below will not succeed, because `main.star` cannot import from non-packages (see [how import works][how-do-kurtosis-imports-work-explanation] for more information):

```python
# Invalid
helpers = import_module("github.com/package-author/package-repo/not-a-package/random-script.star")
```

### Relative Locators
Relative locators like `./helper.star` are allowed as a short alternative to the full absolute locator. However, a relative locator cannot be used to reference files outside the package. In other words, you cannot use a relative locator to reference files above the directory containing the `kurtosis.yml` file.

Suppose we had a [Kurtosis package][packages] like so:

```
/
    package-repo/
        main.star
        src/
            lib.star
```

with a `kurtosis.yml` file like so:

```yaml
name: github.com/package-author/package-repo
```

The `main.star` can refer to the `lib.star` file using either relative or absolute imports:


```python
# valid relative import
lib_via_relative_import = import_module("./src/lib.star")

# valid absolute import
lib_via_absolute_import = import_module("github.com/kurtosis-tech/package-repo/src/lib.star")
```

<!------------------ ONLY LINKS BELOW HERE -------------------->
[packages]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../advanced-concepts/how-do-kurtosis-imports-work.md
[running-private-packages]: ../guides/running-private-packages.md
