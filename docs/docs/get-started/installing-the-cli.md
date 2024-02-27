---
title: Install Kurtosis
id: installing-the-cli
sidebar_label: Install Kurtosis
slug: /install
sidebar_position: 2
toc_max_header: 2
---

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>

The instructions in this guide will walk you through installing the latest version of Kurtosis.

I. Install & Start Docker
-----------------

1. If you don't already have Docker installed, follow the instructions [here][docker-install] to install the Docker application specific to your machine (e.g. Apple Intel, Apple M1, etc.). 
2. Start the Docker daemon (e.g. open Docker Desktop)
3. Verify that Docker is running:
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

Windows users are encouraged to use [Windows Subsystem for Linux (WSL)][windows-susbsystem-for-linux] to use Kurtosis.
If you want to run a native executable, you can download the latest build for your architecture [here](https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/tags).

Or do it using PowerShell:

### 1. Downloading Kurtosis

Download latest Kurtosis package and extract it to current directory.
```bash
Invoke-WebRequest -Uri "https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases/download/REPLACE_VERSION/kurtosis-cli_REPLACE_VERSION_windows_REPLACE_ARCH.tar.gz" -OutFile kurtosis.tar.gz
tar -xvzf kurtosis.tar.gz
```

### 2. Updating the Path to Include Kurtosis

> **Note:** This step needs to be executed in an administrative PowerShell session.

Add Kurtosis to the `Path` environment variable and make a bat script which would take care of running Kurtosis as a cmdlet.
```bash
$currentDir = Get-Location
$systemPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine)
if (-not $systemPath.Contains($currentDir)) {
    $newPath = $systemPath + ";" + $currentDir
    [Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::Machine)
}

$batchContent = @"
@echo off
kurtosis.exe %*
"@
$batchContent | Out-File "$currentDir\kurtosis.bat"
```

### 3. Example Usage

> ⚠️ **Warning**: Ensure you open a new PowerShell window after completing steps 1 and 2 to reflect the updated environment variables.

```bash
kurtosis version
```
</TabItem>

</Tabs>

III. (Optional) Add command-line completion
--------------------------------
Kurtosis supports optional command-line completion to allow completing subcommands and dynamic values (e.g. enclave name during `enclave inspect`). If you'd like to install it, see [these instructions][installing-command-line-completion].

Run the quickstart
------------------
If you're new to Kurtosis, you might like the [quickstart][quickstart] as a good onboarding to get started with Kurtosis.

<!-------------------------- ONLY LINKS BELOW HERE ---------------------------->
[cli-changelog]: ../changelog.md
[metrics-philosophy]: ../advanced-concepts/metrics-philosophy.md
[analytics-disable]: ../cli-reference/analytics-disable.md
[quickstart]: ../get-started/quickstart.md
[installing-command-line-completion]: ../guides/adding-command-line-completion.md
[install-historical-guide]: ../guides/installing-historical-versions.md
[upgrade-guide]: ../guides/upgrading-the-cli.md

[release-artifacts]: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases
[windows-susbsystem-for-linux]: https://learn.microsoft.com/en-us/windows/wsl/
[docker-install]: https://docs.docker.com/get-docker/
