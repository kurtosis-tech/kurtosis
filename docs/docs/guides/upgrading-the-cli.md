---
title: Upgrading Kurtosis
sidebar_label: Upgrading Kurtosis
slug: /upgrade
sidebar_position: 2
---

The instructions in this guide assume you already have Kurtosis installed, and will walk you through upgrading to the latest version of Kurtosis. 

If you're looking to install Kurtosis, [see here][install-guide].

Step One: Verify Breaking Changes
=================================
You can check the version of the CLI you're running with `kurtosis version`. Before upgrading to the latest version, check [the changelog to see if there are any breaking changes][cli-changelog] before proceeding with the steps below to upgrade. 

Step Two: Upgrade The CLI
=========================

<details>
<summary>Homebrew (MacOS)</summary>

```bash
brew update && brew upgrade kurtosis-tech/tap/kurtosis-cli
```
</details>

<details>
<summary>apt (Ubuntu)</summary>

```bash
apt install --only-upgrade kurtosis-cli
```
</details>

<details>
<summary>yum (RHEL)</summary>

```bash
yum upgrade kurtosis-cli
```
</details>

<details>
<summary>deb, rpm, and apk</summary>

Download the appropriate artifact from [the release artifacts page][release-artifacts].
</details>

<details><summary>Windows</summary>

The Kurtosis CLI cannot be installed directly on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.

</details>

Step Three: Restart Engine If Necessary
=======================================
If you upgraded the CLI through a minor version (the `Y` in a `X.Y.Z` version), you may need to restart your Kurtosis engine after the upgrade. 

If this is needed, the Kurtosis CLI will prompt you with an error like so:

```text
The engine server API version that the CLI expects, 1.7.4, doesn't match the running engine server API version, 1.6.8; this would cause broken functionality so you'll need to restart the engine to get the correct version by running 'kurtosis engine restart'
```

The fix is to [restart the engine][kurtosis-engine-restart] like so:

```
kurtosis engine restart
```

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[install-guide]: ./installing-the-cli.md
[cli-changelog]: ../changelog.md
[metrics-philosophy]: ../explanations/metrics-philosophy.md
[quickstart]: ../quickstart.md
[installing-tab-completion]: ./adding-tab-completion.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/

[kurtosis-engine-restart]: ../reference/cli/engine-restart.md
