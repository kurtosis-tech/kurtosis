---
title: Packages
sidebar_label: Packages
---

:::info
Packages are a part of the Kurtosis package system. To read about the package system in detail, [see here][how-do-kurtosis-imports-work-explanation].
:::

<!-- TODO Add more information here when dependencies are specified in the kurtosis.yml -->

A Kurtosis package is a:

- A directory
- Plus all its contents
- That contains [a `kurtosis.yml` file][kurtosis-yml] with the package's name, which will be the [locator][locators] root for the package

Kurtosis packages are [the system by which Starlark scripts can include external resources][how-do-kurtosis-imports-work-explanation].

Note that, when developing locally, the GitHub repo referred to in the package name does not need to exist.

Kurtosis packages are shared simply by pushing to GitHub (e.g. [these are the packages we administer][kurtosis-managed-packages]).

For example, suppose there is a repo called `package-repo` by the author `package-author` whose internal directory structure looks like so:

```
/
    package-repo
        my-package
            kurtosis.yml
            main.star
            helpers
                helpers.star
```

whose `kurtosis.yml` file looked like so:

```yaml
name: github.com/package-author/package-repo/my-package
```

The package would be called `github.com/package-author/package-repo/my-package`. It should get pushed to the `package-repo` repo owned by the `package-author` user on GitHub.

Packages are referenced indirectly, as the [locators][locators] used to specify external resources in a Starlark script will contain the package name where the resource lives.

For example:

```python
helpers = import_module("github.com/package-author/package-repo/my-package/helpers/helpers.star")
```

would be used to import the `helpers.star` file into a Starlark script.

<!-- TODO Update this when dependencies are done in the kurtosis.yml file, which would happen at dependency resolution time -->
The Kurtosis engine will automatically download dependency packages from GitHub when running a Starlark script.

### Runnable Packages
A Kurtosis package that has a `main.star` file next to its `kurtosis.yml` file is called a "runnable package". The `main.star` file of a runnable package must have a `run(plan)` method like so:

```python
def run(plan):
    print("Hello, world.")
```

:::info
More on the `plan` parameter [here][plan].
:::

Runnable packages can be executed from the CLI in one of three ways:

```bash
# OPTION 1: Point to a directory with a `kurtosis.yml` and `main.star` on local filesystem
kurtosis run /path/to/runnable/package/root
```

```bash
# OPTION 2: Point to a `kurtosis.yml` on the local filesystem with a `main.star` next to it on local fileesystem
kurtosis run /path/to/runnable/package/root/kurtosis.yml
```

```bash
# OPTION 3: Pass in a remote package name to run from GitHub
kurtosis run github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml
```

:::tip
If you want to run a non-main branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

All these will call the `run(plan)` function of the package's `main.star`.

### Arguments
Kurtosis packages can be parameterized with arguments by adding an `args` parameter to the `run` function. Read more about package arguments [here][args-reference].

<!-------------------- ONLY LINKS BELOW HERE -------------------------->
[kurtosis-yml]: ./kurtosis-yml.md
[locators]: ./locators.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=package+in%3Aname&type=all&language=&sort=
[how-do-kurtosis-imports-work-explanation]: ../explanations/how-do-kurtosis-imports-work.md
[plan]: ./plan.md
[args-reference]: ./args.md
