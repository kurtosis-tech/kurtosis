---
title: Installing Historical Versions
sidebar_label: Installing Historical Versions
slug: /install-historical
sidebar_position: 3
---

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>

Occasionally, using historical versions of Kurtosis is necessary. For example, when working with a [Starlark](../concepts-reference/starlark.md) Kurtosis package that was initially developed with an older version of Kurtosis, we might want rollback our Kurtosis version to ensure the version of Kurtosis we are running is compatible with the Kurtosis package.

The instructions in this guide will walk you through installing and using a historical version of Kurtosis. To see what versions are available, reference our [changelog](../changelog.md).

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
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
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
baseurl=https://yum.fury.io/kurtosis-tech/
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

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[install-kurtosis]: ./installing-the-cli.md
[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
