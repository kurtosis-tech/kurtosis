---
title: Installing Kurtosis
sidebar_label: Installing Kurtosis
slug: /install
sidebar_position: 1
---

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>


The instructions in this guide will walk you through installing the latest version of Kurtosis. 

If you already have Kurtosis installed and you're looking to upgrade to latest, [see here][upgrade-guide].

If you're looking to install a historical version instead, [see here][install-historical-guide].

I. Install & Start Docker
-----------------

1. If you don't already have Docker installed, follow the instructions [here][docker-install] to install the Docker application specific to your machine (e.g. Apple Intel, Apple M1, etc.). 
1. Start the Docker daemon (e.g. open Docker Desktop)
1. Verify that Docker is running:
   ```bash
   docker image ls
   ```

II. Install the CLI
-------------------------

<Tabs groupId="install-methods">
<TabItem value="homebrew" label="brew (MacOS)">

```
brew install kurtosis-tech/tap/kurtosis-cli
```

:::info
Homebrew might warn you that your Xcode is outdated or missing entirely. [This is a Homebrew requirement](https://docs.brew.sh/Installation), and has nothing to do with Kurtosis (which ships as prebuilt binaries). 

To install or update your Xcode, run:

```bash
xcode-select --install
```
:::

</TabItem>
<TabItem value="apt" label="apt (Ubuntu)">

```bash
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```

</TabItem>
<TabItem value="yum" label="yum (RHEL)">

```bash
echo '[kurtosis]
name=Kurtosis
baseurl=https://yum.fury.io/kurtosis-tech/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum install kurtosis-cli
```

</TabItem>
<TabItem value="other-linux" label="deb, rpm, and apk">

Download the appropriate artifact from [the release artifacts page][release-artifacts].

</TabItem>

<TabItem value="windows" label="Windows">

The Kurtosis CLI cannot be installed directly on Windows. Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.

</TabItem>

</Tabs>

III. (Optional) Add command-line completion
--------------------------------
Kurtosis supports command-line completion to allow completing subcommands and dynamic values (e.g. enclave name during `enclave inspect`). This isn't required, but we believe it significantly enhances the Kurtosis experience for those who are using . If you'd like to install it, see [these instructions][installing-command-line-completion].

Run the quickstart
------------------
If you're new to Kurtosis, you might like the [quickstart][quickstart] as a good onboarding to get started with Kurtosis.

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[cli-changelog]: ../changelog.md
[metrics-philosophy]: ../explanations/metrics-philosophy.md
[analytics-disable]: ../cli-reference/analytics-disable.md
[quickstart]: ../quickstart.md
[installing-command-line-completion]: ./adding-command-line-completion.md
[install-historical-guide]: ./installing-historical-versions.md
[upgrade-guide]: ./upgrading-the-cli.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/
[docker-install]: https://docs.docker.com/get-docker/
