---
id: index
title: CLI Introduction
sidebar_label: Introduction
slug: /cli
sidebar_position: 1
---

The Kurtosis CLI is a Go CLI wrapped around the Kurtosis Go [client library][client-library-reference]. This section will go through the most common Kurtosis CLI commands and some useful tips on getting started. If you have not already done so, the CLI can be installed by following the instructions [here][installing-the-cli].


:::tip
Kurtosis supports command-line completion; we recommend [installing it][adding-command-line-completion] for the best experience.
:::

### Configuration file path
To locate where the Kurtosis configuration file is on your machine, simply use:

```bash
kurtosis config path
```
to print out the file path of the `kurtosis-config.yml` file.

### Get the CLI version
The version of the CLI and the currently-running engine can be printed with the following:

```
kurtosis version
```

### Global Flags
The Kurtosis CLI supports two global flags - `help` and `cli-log-level`. These flags can be used with any Kurtosis CLI command.

#### -h or --help
This flag prints the help text for all commands and subcommands. You can use this at any time to see information on the command you're trying to run. For example:
```
kurtosis service -h
```
<details>
    <summary>Example Output of the above command</summary>

```bash
Manage services

Usage:
  kurtosis service [command]

Available Commands:
  add         Adds a service to an enclave
  logs        Get service logs
  rm          Removes a service from an enclave
  shell       Gets a shell on a service

Flags:
  -h, --help   help for service

Global Flags:
      --cli-log-level string   Sets the level that the CLI will log at (panic|fatal|error|warning|info|debug|trace) (default "info")

Use "kurtosis service [command] --help" for more information about a command.
```
</details>


#### cli-log-level
This flag sets the level of details that the Kurtosis CLI will print logs with - by default it only logs `info` level logs to the CLI. The following other log levels are supported by Kurtosis -
`panic|fatal|error|warning|info|debug|trace`. For example, logs with error level can be printed using the command below:-

```
kurtosis run --cli-level-log debug github.com/package-author/package-repo 
```

<details>
    <summary>Example Output of the above command</summary>

```bash
DEBU[2023-04-03T12:54:00-04:00] Instantiating a context aware backend with no remote backend config ends up returninga regular local Docker backend. 
INFO[2023-04-03T12:54:00-04:00] No Kurtosis engine was found; attempting to start one... 
DEBU[2023-04-03T12:54:00-04:00] Metrics user id filepath: '' 
INFO[2023-04-03T12:54:00-04:00] Pulling image 'kurtosistech/engine:0.73.0'... 
DEBU[2023-04-03T12:54:00-04:00] Binds: [/var/run/docker.sock:/var/run/docker.sock] 
DEBU[2023-04-03T12:54:00-04:00] Created container with ID 'b9c8f6509ebe7831a96a926e0514f049884b30a8ff4359cd06d9592464d7f017' from image 'kurtosistech/engine:0.73.0' 
DEBU[2023-04-03T12:54:01-04:00] Netstat availability-waiting command '[ -n "$(netstat -anp tcp | grep LISTEN | grep 9710)" ]' returned without a Docker error, but exited with non-0 exit code '1' and logs: 
INFO[2023-04-03T12:54:02-04:00] Successfully started Kurtosis engine         
DEBU[2023-04-03T12:54:02-04:00] Kurtosis Portal daemon is currently not reachable. If Kurtosis is being used ona local-only context, this is fine as Portal is not required for local-only contexts. 
INFO[2023-04-03T12:54:02-04:00] Creating a new enclave for Starlark to run inside... 
INFO[2023-04-03T12:54:04-04:00] Enclave 'murky-volcano' created successfully 
INFO[2023-04-03T12:54:04-04:00] Executing Starlark package at '' as the passed argument '' looks like a directory 
INFO[2023-04-03T12:54:04-04:00] Compressing package '' at '' for upload 
INFO[2023-04-03T12:54:04-04:00] Uploading and executing package '' 

> print msg={"key": "value"}
{"key": "value"}

Starlark code successfully run. No output was returned.
DEBU[2023-04-03T12:54:04-04:00] Successfully reached the end of the response stream. Closing. 
DEBU[2023-04-03T12:54:04-04:00] Current context is local, not mapping enclave service ports 
INFO[2023-04-03T12:54:04-04:00] ====================================================== 
INFO[2023-04-03T12:54:04-04:00] ||          Created enclave: murky-volcano          || 
INFO[2023-04-03T12:54:04-04:00] ====================================================== 
Name:            murky-volcano
UUID:            f2fa01a0293f
Status:          RUNNING
Creation Time:   Mon, 03 Apr 2023 12:54:02 EDT

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID   Name   Ports   Status
```
</details>


:::info
Users can use the `debug` `--cli-log-level` flag, , as shown above, to display the entire stack trace to the CLI. By default the entire stack trace is saved to the `kurtosis-cli.log` file. 

The sample error stack-trace that can be seen on the cli after `debug` level is shown below:

```bash
DEBU[2023-04-03T12:58:03-04:00] Cluster setting filepath: '' 
DEBU[2023-04-03T12:58:03-04:00] Kurtosis config YAML filepath: '' 
DEBU[2023-04-03T12:58:03-04:00] Loaded Kurtosis Config  &{overrides:0x1400000e510 shouldSendMetrics:true clusters:map[docker:0x14000097680 minikube:0x140000976b0]} 
DEBU[2023-04-03T12:58:03-04:00] Instantiating a context aware backend with no remote backend config ends up returninga regular local Docker backend. 
Error:  An error occurred validating arg ''
 --- at /root/project/cli/cli/command_framework/lowlevel/lowlevel_kurtosis_command.go:290 (LowlevelKurtosisCommand.MustGetCobraCommand.func2) ---
Caused by: Error reading filepath_or_dirpath ''
 --- at /root/project/cli/cli/command_framework/highlevel/file_system_path_arg/file_system_path_arg.go:109 (getValidationFunc.func1) ---
Caused by: stat ../../../per/other/submodul/: no such file or directory
```
:::

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[adding-command-line-completion]: ../guides/adding-command-line-completion.md
[installing-the-cli]: ../guides/installing-the-cli.md
[client-library-reference]: ../client-libs-reference.md
