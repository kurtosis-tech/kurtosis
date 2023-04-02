---
title: kurtosis.yml
sidebar_label: kurtosis.yml
---

:::info
The `kurtosis.yml` is part of the Kurtosis package system. To read about the package system in detail, [see here][how-do-kurtosis-imports-work-explanation].
:::

The `kurtosis.yml` file is a manifest file necessary to turn a directory into [a Kurtosis package][package]. This is the spec for the `kurtosis.yml`:

<!-- TODO UPDATE THIS WHEN DEPENDENCIES GO HERE -->

```yaml
# The locator naming this package.
name: github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml
```

:::info
The URL-like string is **not** a Github URL! This string is called a [locator][locators-reference], and it is inspired by the Golang module system.
:::

Example usage:

If the `kurtosis.yml` is in the repository root:

```yaml
name: github.com/author/package-repo
```

If the `kurtosis.yml` is in a directory other than repository root:
```yaml
name: github.com/author/package-repo/path/to/directory-with-kurtosis.yml
```

:::info
The key take away is that `/path/to/directory-with-kurtosis.yml` only needs to be provided if `kurtosis.yml` is in a subdirectory rather than the repo root.
:::

<!----------------------- ONLY LINKS BELOW HERE ----------------------------->
[package]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../explanations/how-do-kurtosis-imports-work.md
[locators-reference]: ./locators.md
