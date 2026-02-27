---
title: Installing Old Versions
sidebar_label: Installing Old Versions
slug: /install-historical
sidebar_position: 4
---

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>

Occasionally, using older versions of Kurtosis is necessary. For example, when working with a [Starlark](../advanced-concepts/starlark.md) Kurtosis package that was initially developed with an older version of Kurtosis, we might want rollback our Kurtosis version to ensure the version of Kurtosis we are running is compatible with the Kurtosis package.

The instructions in this guide will walk you through installing and using an older version of Kurtosis. To see what versions are available, reference our [changelog](../changelog.md).

If you're looking to install the latest version of Kurtosis, [see here][install-kurtosis].


<Tabs groupId="install-methods">
<TabItem value="homebrew" label="brew (MacOS)">

1. Uninstall your current version of Kurtosis
    ```
    brew uninstall kurtosis-tech/tap/kurtosis-cli
    ```
   
2. Install an earlier version of Kurtosis (eg. `0.68.6`)
   ```
   brew install kurtosis-tech/tap/kurtosis-cli@<version>
   ```

</TabItem>
<TabItem value="apt" label="apt (Ubuntu)">

:::caution

If you already have `kurtosis-cli` package installed, we recommend uninstalling it first using:

```bash
sudo apt remove kurtosis-cli
```

:::

```bash
echo "deb [trusted=yes] https://kurtosis-tech.github.io/kurtosis-cli-release-artifacts/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt remove kurtosis-cli
sudo apt install kurtosis-cli=<version> -V
```

</TabItem>
<TabItem value="yum" label="yum (RHEL)">

:::caution

If you already have `kurtosis-cli` package installed, we recommend uninstalling it first using:

```bash
sudo yum remove kurtosis-cli
```

:::

```bash
echo '[kurtosis]
name=Kurtosis
baseurl=https://kurtosis-tech.github.io/kurtosis-cli-release-artifacts/rpm
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum remove kurtosis-cli
sudo yum install kurtosis-cli-<version>
```

</TabItem>
<TabItem value="other-linux" label="deb, rpm, and apk">

Download the appropriate artifact from [the release artifacts page][release-artifacts].

</TabItem>

<TabItem value="windows" label="Windows">

The Kurtosis CLI cannot be installed directly on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)](https://learn.microsoft.com/en-us/windows/wsl/install) to use Kurtosis.

</TabItem>

</Tabs>

:::tip
In order to upgrade Kurtosis to another version *after you've performed a downgrade (i.e. installed a historical version)*, you must first uninstall the version of Kurtosis you've installed and re-install Kurtosis. When using Homebrew, the workflow will be (replacing `HISTORICAL-VERSION` with the historical version you have installed):
1. `brew uninstall brew uninstall kurtosis-tech/tap/kurtosis-cli@HISTORICAL-VERSION`
2. `brew install kurtosis-tech/tap/kurtosis-cli` for upgrading to the latest version or `brew install kurtosis-tech/tap/kurtosis-cli@TARGET-VERSION` for upgrading to a specific version
3. `kurtosis engine restart`
:::

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[install-kurtosis]: ../get-started/installing-the-cli.md
[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
