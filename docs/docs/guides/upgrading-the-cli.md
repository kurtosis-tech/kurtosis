---
title: Upgrading Kurtosis
sidebar_label: Upgrading Kurtosis
slug: /upgrade
sidebar_position: 2
---

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>

The instructions in this guide assume you already have Kurtosis installed, and will walk you through upgrading to the latest version of Kurtosis. 

If you're looking to install Kurtosis, [see here][install-guide].

I. Check breaking changes
---------------------------------
You can check the version of the CLI you're running with `kurtosis version`. Before upgrading to the latest version, check [the changelog to see if there are any breaking changes][cli-changelog] before proceeding with the steps below to upgrade. 

II. Upgrade the CLI
-------------------------

<Tabs groupId="install-methods">
<TabItem value="homebrew" label="brew (MacOS)">

```bash
brew update && brew upgrade kurtosis-tech/tap/kurtosis-cli
```

</TabItem>
<TabItem value="apt" label="apt (Ubuntu)">

```bash
apt install --only-upgrade kurtosis-cli
```

</TabItem>
<TabItem value="yum" label="yum (RHEL)">

```bash
yum upgrade kurtosis-cli
```

</TabItem>
<TabItem value="other-linux" label="deb, rpm, and apk">

Download the appropriate artifact from [the release artifacts page][release-artifacts].

</TabItem>

<TabItem value="windows" label="Windows">

The Kurtosis CLI cannot be installed directly on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.

</TabItem>

</Tabs>

III. Restart the engine
-----------------------
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
[installing-command-line-completion]: ./adding-command-line-completion.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/

[kurtosis-engine-restart]: ../cli-reference/engine-restart.md
