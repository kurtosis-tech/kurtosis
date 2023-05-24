---
title: Locators
sidebar_label: Locators
---

:::info
Locators are a part of the Kurtosis package system. To read about the package system in detail, [see here][how-do-kurtosis-imports-work-explanation].
:::

A locator is a URL-like string used to locate a resource inside [a Kurtosis package][packages]. For example, this locator:

```
github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml/some-file.star
```

references a file inside a GitHub repo called `package-repo`, owned by `package-author`, that lives at the path `/path/to/directory-with-kurtosis.yml/some-file.star` relative to the root of the repo.


Locators are used for identifying resources that will be used inside a Starlark script - namely by [`import_module`](../starlark-reference/import-module.md) and [`read_file`](../starlark-reference/read-file.md).

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
Only locators pointing to public GitHub repositories are currently allowed.
:::

Any Starlark script that wishes to use external resources must be
a part of a [Kurtosis package][packages].

All locators are absolute; "relative" locators do not exist. For a Starlark script to reference a local file (i.e. one that lives next to in the filesystem), the Starlark script must use the name of the package that it lives inside.

For example, suppose we had a [Kurtosis package][packages] like so:

```
/
    package-repo
        my-package 
            kurtosis.yml
            main.star
            helpers
                random-script.star
        not-a-package
            random-script.star
```

with a `kurtosis.yml` file like so:

```yaml
name: github.com/package-author/package-repo/my-package
```

The `main.star` file would import the `random-script.star` from the `helpers` subdirectory of `my-package` like so:

```python
helpers = import_module("github.com/package-author/package-repo/my-package/helpers/random-script.star")
```

The import statement below will not succeed, this is because `main.star` cannot import from non-packages.
(see [how import works][how-do-kurtosis-imports-work-explanation] for more information)

```python
helpers = import_module("github.com/package-author/package-repo/not-a-package/random-script.star")
```

<!------------------ ONLY LINKS BELOW HERE -------------------->
[packages]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../explanations/how-do-kurtosis-imports-work.md
