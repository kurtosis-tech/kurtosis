---
title: Resource Identifier
sidebar_label: Resource Identifier
---

Kurtosis has multiple ways to identify a given resource within Kurtosis. These include UUIDs, shortened UUIDs and names. Together these are called resource identifiers.

A resource identifier as mentioned above is the union of the following -

- UUID - a UUID or a Universally Unique Identifier within Kurtosis is a 32 character-long hex-encoded string generated using [UUID v4][uuidv4]. Kurtosis automatically assigns a UUID to resources.
- Shortened UUID - A shortened UUID is the first 12 characters of a UUID. As the shortened UUID is just 12 characters, it isn't guaranteed to be unique. In case of conflicts, Kurtosis tells the user about the ambiguity and provides
a list of matching full UUIDs.
- Name - A name is what a user gives to the underlying resource. For some resources Kurtosis automatically generates the name if the user doesn't specify it. Note a name is only point-in-time unique; you can have the same name identifying different resources over time. In case of conflicts, Kurtosis tells the user about the ambiguity and provides a list of matching full UUIDs.

For example, let's assume an enclave was created with the name `winter-sun`. If you run `kurtosis enclave inspect winter-sun` you would get something like -

```
UUID:                                 edfdbf5766f6
Enclave Name:                         winter-snow
Enclave Status:                       RUNNING
Creation Time:                        Tue, 24 Jan 2023 16:45:13 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:54433
API Container Host GRPC Proxy Port:   127.0.0.1:54434

========================================== User Services ==========================================
UUID   Name   Ports   Status
```

Notice how Kurtosis shows the shortened UUID by default. All CLI commands show
shortened UUIDs by default; if you want to see full UUIDs you can use the `--full-uuids` flag with the command. Rerunning the above command with `--full-uuids` you'd get

```
UUID:                                 edfdbf5766f64a649efca11f51ebb4c1
Enclave Name:                         winter-snow
Enclave Status:                       RUNNING
Creation Time:                        Tue, 24 Jan 2023 16:45:13 GMT
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:54433
API Container Host GRPC Proxy Port:   127.0.0.1:54434

========================================== User Services ==========================================
UUID   Name   Ports   Status
```

Whenever a user is querying Kurtosis using the CLI & SDK, the user can use any of UUID, shortened UUID & name. The CLI informs the user about this support by calling the relevant argument `resource-identifier`, like below

```
Usage:
  kurtosis enclave inspect [flags] enclave-identifier

Flags:
      --full-uuids   If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.
  -h, --help        help for inspect

Global Flags:
      --cli-log-level string   Sets the level that the CLI will log at (panic|fatal|error|warning|info|debug|trace) (default "info")
```

Resource Identifiers are supported for the following resources inside of Kurtosis

- [Files Artifacts][files-artifacts]
- [Services][services]
- [Enclaves][enclaves]

<!------------------ ONLY LINKS BELOW HERE -------------------->
[uuidv4]: https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)
[files-artifacts]: ./files-artifacts.md
[services]: ./glossary.md#user-service
[enclaves]: ./glossary.md#enclave
