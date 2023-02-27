---
title: Installing The CLI
sidebar_label: Installing The CLI
slug: /install
sidebar_position: 1
---

Interacting with Kurtosis is done via [a CLI](../reference/cli/cli.md). The instructions below will walk you through installing it.

:::tip
Kurtosis supports tab completion, and we strongly recommend [installing it][installing-tab-completion] after you install the CLI.
:::

<details>
<summary>Homebrew</summary>

```
brew install kurtosis-tech/tap/kurtosis-cli
```

NOTE: Homebrew might warn you that your Xcode is outdated, like so:

```
Error: Your Xcode (11.5) is too outdated.
Please update to Xcode 12.5 (or delete it).
```

[This is a Homebrew requirement](https://docs.brew.sh/Installation), and has nothing to do with Kurtosis (which ships as prebuilt binaries). To update your Xcode, run:

```
xcode-select --install
```
</details>

<details>
<summary>apt</summary>

```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```
</details>

<details>
<summary>yum</summary>

```
echo '[kurtosis]
name=Kurtosis
baseurl=https://yum.fury.io/kurtosis-tech/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum install kurtosis-cli
```
</details>

<details>
<summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>

:::info
Kurtosis CLI cannot be installed on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.
:::

Once you're done, [the quickstart is a great place to get started][quickstart].

Analytics
----------

On installation Kurtosis enables anonymized [analytics][metrics-philosophy] by default. In case you want to disable
it you can run `kurtosis analytics disable`.

Upgrading
---------
You can check the version of the CLI you're running with `kurtosis version`. To upgrade to latest, check [the changelog to see if there are any breaking changes][cli-changelog] and follow the steps below. 

NOTE: if you're upgrading the CLI's minor version (the `Y` in a `X.Y.Z` version), you may need to restart your Kurtosis engine after the upgrade. If this is needed, the Kurtosis CLI will prompt you with an error like so:
```
The engine server API version that the CLI expects, 1.7.4, doesn't match the running engine server API version, 1.6.8; this would cause broken functionality so you'll need to restart the engine to get the correct version by running 'kurtosis engine restart'
```
The fix is to restart the engine like so:
```
kurtosis engine restart
```

<details>
<summary>Homebrew</summary>

```
brew upgrade kurtosis-tech/tap/kurtosis-cli
```
</details>

<details>
<summary>apt</summary>

```
apt install --only-upgrade kurtosis-cli
```
</details>

<details>
<summary>yum</summary>

```
yum upgrade kurtosis-cli
```
</details>

<details>
<summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[cli-changelog]: ../changelog.md
[metrics-philosophy]: ../explanations/metrics-philosophy.md
[quickstart]: ../quickstart.md
[installing-tab-completion]: ./adding-tab-completion.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/
