---
title: Installing, Upgrading & Configuring The CLI
sidebar_label: Installing, Upgrading & Configuring The CLI
slug: /install
sidebar_position: 1
---

The instructions in this guide will walk you through installing, upgrading, & configuring the CLI. 

:::tip
Kurtosis supports tab completion, and we strongly recommend [installing it][installing-tab-completion] after you finish installing the CLI.
:::

Installing the CLI
------------------

There are a few ways to install the CLI and the below section will guide you through each of those ways. Once you're done installing the CLI, the [quickstart][quickstart] is a great place to get started.

<details><summary>MacOS (Homebrew)</summary>

To install on MacOS (Homebrew):

```
brew install kurtosis-tech/tap/kurtosis-cli
```

NOTE: Homebrew might warn you that your Xcode is outdated or missing entirely. [This is a Homebrew requirement](https://docs.brew.sh/Installation), and has nothing to do with Kurtosis (which ships as prebuilt binaries). To install or update your Xcode, run:

```
xcode-select --install
```

</details>

<details><summary>apt</summary>

To install on Ubuntu OS using apt:

```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```
</details>

<details><summary>yum</summary>

To install on RPM-based Linux systems:

```
echo '[kurtosis]
name=Kurtosis
baseurl=https://yum.fury.io/kurtosis-tech/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum install kurtosis-cli
```
</details>

<details><summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>

:::info
The Kurtosis CLI cannot be installed directly on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.
:::

Upgrading the CLI
-----------------
You can check the version of the CLI you're running on by using the command: `kurtosis version`. Before upgrading to the latest version, we recommend checking [the changelog to see if there are any breaking changes][cli-changelog] before proceeding with the steps below to upgrade.

:::tip
if you're upgrading the CLI's minor version (the `Y` in a `X.Y.Z` version), you may need to restart your Kurtosis engine after the upgrade. If this is needed, the Kurtosis CLI will prompt you with an error like so:
```
The engine server API version that the CLI expects, 1.7.4, doesn't match the running engine server API version, 1.6.8; this would cause broken functionality so you'll need to restart the engine to get the correct version by running 'kurtosis engine restart'
```
The fix is to restart the engine like so:
```
kurtosis engine restart
```
:::

<details><summary>Homebrew</summary>

To upgrade the CLI on MacOS (Homebrew):

```
brew upgrade kurtosis-tech/tap/kurtosis-cli
```

If you encounter issues with upgrading the CLI using Homebrew, try the following command to update and upgrade Homebrew itself before upgrading the CLI:
```
brew update && brew upgrade
```

</details>

<details><summary>apt</summary>

To upgrade the CLI on Ubuntu OS using apt:

```
apt install --only-upgrade kurtosis-cli
```
</details>

<details><summary>yum</summary>

To upgrade the CLI on RPM-based Linux systems:

```
yum upgrade kurtosis-cli
```
</details>

<details><summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>


Configuring Analytics
---------------------

On installation, Kurtosis enables anonymized analytics by default. In case you want to disable it, you can run: `kurtosis analytics disable` to [disable the sending of product analytics metrics][analytics-disable]. 

Read more about why and how we collect product analytics metrics [here][metrics-philosophy].


<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[cli-changelog]: ../changelog.md
[metrics-philosophy]: ../explanations/metrics-philosophy.md
[analytics-disable]: ../reference/cli/analytics-disable.md
[quickstart]: ../quickstart.md
[installing-tab-completion]: ./adding-tab-completion.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/
