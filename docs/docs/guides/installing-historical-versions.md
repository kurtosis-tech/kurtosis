---
title: Installing Historical Versions of Kurtosis CLI
sidebar_label: Installing Historical Versions
slug: /install-historical
sidebar_position: 2
---

Occasionally, using historical versions of Kurtosis is necessary. For example, when working with a [Starlark](../explanations/starlark.md) Kurtosis package that was initially developed with an older version of Kurtosis, we might want rollback our Kurtosis version to ensure the version of Kurtosis we are running is compatible with the Kurtosis package.

The instructions below walk you through installing and using a historical version of Kurtosis. To see what versions are available, reference our [changelog](../changelog.md)!

<details>
<summary>Homebrew</summary>

1. Uninstall your current version of Kurtosis
    ```
    brew uninstall kurtosis-tech/tap/kurtosis-cli
    ```

2. Install an earlier version of Kurtosis
   ```
   brew install kurtosis-tech/tap/kurtosis-cli@<version>
   ```

</details>

<details>
<summary>apt</summary>
:::caution

If you already have `kurtosis-cli` package installed, we recommend uninstalling it first using `sudo apt remove kurtosis-cli`.

:::
```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt remove kurtosis-cli
sudo apt install kurtosis-cli=<version> -V
```
</details>
<details>
<summary>yum</summary>

:::caution

If you already have `kurtosis-cli` package installed, we recommend uninstalling it first using `sudo yum remove kurtosis-cli`.

:::
```
echo '[kurtosis]
name=Kurtosis
baseurl=https://yum.fury.io/kurtosis-tech/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum remove kurtosis-cli
sudo yum install kurtosis-cli-<version>
```
</details>

<details>
<summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
