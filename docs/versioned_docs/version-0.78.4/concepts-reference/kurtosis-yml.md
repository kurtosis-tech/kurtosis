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

Example usage:

if kurtosis.yml is in the repository root:
```yaml
name: github.com/author/package-repo
```

if kurtosis.yml is in a directory other than repository root:
```yaml
name: github.com/author/package-repo/path/to/directory-with-kurtosis.yml
```

:::info
The key take away is that `/path/to/directory-with-kurtosis.yml` only needs to be provided if `kurtosis.yml` is not present in the repository's root.
:::

<!----------------------- ONLY LINKS BELOW HERE ----------------------------->
[package]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../explanations/how-do-kurtosis-imports-work.md