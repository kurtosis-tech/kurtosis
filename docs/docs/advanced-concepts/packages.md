---
title: Packages
sidebar_label: Packages
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

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
    package-repo/
        my-package/
            kurtosis.yml
            main.star
            helpers/
                helpers.star
```

whose `kurtosis.yml` file looks like so:

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
    plan.print("Hello, world.")
```

:::info
More on the `plan` parameter [here][plan].
:::

Runnable packages can be called through the `kurtosis run` function of the CLI:

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

### Parameterization
Kurtosis packages can accept parameters, allowing their behaviour to change. 

To make your package take in arguments, first add extra parameters to your package's `run` function:

From this...

```python
def run(plan):
```

...to this:

```python
# Parameters without a default value are required; parameter with a default value are optional
def run(plan, some_parameter, some_other_parameter = "Default value"):
```

:::warning
You may come across an old style of package parameterization where the `run` function takes a single `args` variable containing all the package's parameters, like so:

```python
# OLD STYLE - DO NOT USE
def run(plan, args):
```

This method is now deprecated, and will be removed in the future.
:::

Consumers of your package can then pass in these parameters to configure your package:

<Tabs>
<TabItem value="cli" label="CLI" default>

```bash
kurtosis run github.com/YOUR-USER/YOUR-REPO '{"some_parameter": 5, "some_other_parameter": "New value"}'
```
For detailed instructions on passing arguments via the CLI, see the ["Arguments" section of the `kurtosis run` documentation][kurtosis-run-arguments].

</TabItem>
<TabItem value="starlark" label="Starlark">

```python
your_package = import_module("github.com/YOUR-USER/YOUR-REPO/main.star")

def run(plan):
    your_package.run(plan, some_parameter = 5, some_other_parameter = "New value")
```

</TabItem>
</Tabs>

### Package Icons

Once your package is [published], it will appear in the Kurtosis package catalog found in the web UI. By default a plain
icon is shown - but you can select your own icon by including a `kurtosis-package-icon.png` file alongside your
`kurtosis.yml` file. The image should be square and at least `150px x 150px`. 

<!-------------------- ONLY LINKS BELOW HERE -------------------------->
[kurtosis-yml]: ./kurtosis-yml.md
[locators]: ./locators.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=package+in%3Aname&type=all&language=&sort=
[how-do-kurtosis-imports-work-explanation]: ../advanced-concepts/how-do-kurtosis-imports-work.md
[plan]: ./plan.md
[kurtosis-run-arguments]: ../cli-reference/run.md#arguments
[published]: /quickstart-write-a-package#publishing-your-kurtosis-package-for-others-to-use