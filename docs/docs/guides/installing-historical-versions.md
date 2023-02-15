---
title: Installing Historical Versions of Kurtosis CLI
sidebar_label: Installing Historical Versions
slug: /install-historical
sidebar_position: 1
---

Occasionally, using historical versions of Kurtosis CLI is necessary. For example, when working with a [Starlark](../explanations/starlark.md) Kurtosis package that was initially developed with an older version of Kurtosis, we might want rollback our Kurtosis CLI version to ensure the version of Kurtosis we are running is compatible with the Kurtosis package.

The instructions below walk you through installing and using a historical version of Kurtosis. To see what versions are available, reference our [changelog](../changelog.md) (ex. `0.66.0, 0.66.5`)!

<details>
<summary>Homebrew</summary>

1. Uninstall your current version of Kurtosis CLI
    ```
    brew uninstall kurtosis-tech/tap/kurtosis-cli
    ```

1. Navigate to our public [homebrew tap repository](https://github.com/kurtosis-tech/homebrew-tap).

1. Click on `Releases` and navigate to the release of the version you'd like to install.

1. From that release, click on the `Assets` dropdown. There should appear a list of [homebrew bottles](https://docs.brew.sh/Bottles). Download the bottle associated with your OS and architecture.

1. Install that version of Kurtosis CLI straight from the bottle using this command:
    ```
    brew install -f <bottle-filename>.tar.gz
    ```

1. Doublecheck that you've successfully installed the desired version using this command:
   ``` 
   kurtosis version
   ```
   For example, if we installed version `0.66.1`, the following output should be displayed:
   ```
   WARN[2023-02-14T10:39:14-05:00] You are running an old version of the Kurtosis CLI; we suggest you to update it to the latest version, '0.66.4'
   WARN[2023-02-14T10:39:14-05:00] You can manually upgrade the CLI tool following these instructions: https://docs.kurtosis.com/install#upgrading
   0.66.1
   ```

:::caution

If you are on mac, you might receive a pop up along the lines of `"kurtosis" can't be opened because the developer cannot be verified.` To allow Kurtosis to run, navigate to `System Preferences -> Security & Privacy` and click `Allow Anyways` when prompted regarding `kurtosis`. Then, attempt to run `kurtosis` once more and click `Open` on the subsequent dialog box. More information on allowing mac apps from unidentified developers can be found [here](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac).

:::


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

