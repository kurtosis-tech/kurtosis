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
name: github.com/packageAuthor/packageRepoName
```

<!-- TODO delete this when packages can live in subdirectories -->
:::caution
Only packages at the root of the repo are currently supported (i.e. where the `kurtosis.yml` is at the root of the repo). Packages in subdirectories will be supported soon.
:::

If you're only running the package locally, the `packageAuthor` and `packageRepoName` in the `name` field can be anything. In other words, when running a package locally, the GitHub repository does not in fact need to exist. Once pushed to Github though, `packageAuthor` and `packageRepoName` must match the Github repo's author and name.

<!----------------------- ONLY LINKS BELOW HERE ----------------------------->
[package]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../explanations/how-do-kurtosis-imports-work.md
