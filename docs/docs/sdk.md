---
title: SDK
sidebar_label: SDK
slug: /SDK
toc_min_heading_level: 2
toc_max_heading_level: 2
---

Interactions with Kurtosis happen via API. To facilitate interaction with Kurtosis, we provide [client libraries][kurtosis-client-libs] for interacting with the Kurtosis API. These can be used to, for example, write Kurtosis tests using your test framework of choice.

This page documents the objects and functions in the client libraries.

:::tip
The sidebar on the right can be used to quickly navigate classes.
:::

KurtosisContext
---------------
A connection to a Kurtosis engine, used for manipulating enclaves.

### `createEnclave(String enclaveName, boolean isSubnetworkingEnabled) -> [EnclaveContext][enclavecontext] enclaveContext`
Creates a new Kurtosis enclave using the given parameters.

**Args**
* `enclaveName`: The name to give the new enclave.
* `isSubnetworkingEnabled`: If set to true, the enclave will be set up to allow for subnetworking. This will make service addition & removal take slightly longer, but will enable [subnetworking](./concepts-reference/subnetworks.md) and allow the use of [`Plan.set_connection`](./starlark-reference/plan.md#set_connection) in the Starlark scripts you run.

**Returns**
* `enclaveContext`: An [EnclaveContext][enclavecontext] object representing the new enclave.

### `getEnclaveContext(String enclaveIdentifier) -> [EnclaveContext][enclavecontext] enclaveContext`
Gets the [EnclaveContext][enclavecontext] object for the given enclave ID.

**Args**
* `enclaveIdentifier`: The [identifier][identifier] of the enclave.

**Returns**
* `enclaveContext`: The [EnclaveContext][enclavecontext] representation of the enclave.

### `getEnclaves() -> Enclaves enclaves`
Gets the enclaves that the Kurtosis engine knows about.

**Returns**
* `enclaves`: The [Enclaves][enclaves] representation of all enclaves that Kurtosis the engine knows about.

### `getEnclave(String enclaveIdentifier) -> EnclaveInfo enclaveInfo`
Gets information about the enclave for the given identifier

**Args**
* `enclaveIdentifier`: The [identifier][identifier] of the enclave.

**Returns**
* `enclaves`: The [EnclaveInfo][enclaveinfo] object representing the enclave

### `stopEnclave(String enclaveIdentifier)`
Stops the enclave with the given [identifier][identifier], but doesn't destroy the enclave objects (containers, networks, etc.) so they can be further examined.

**NOTE:** Any [EnclaveContext][enclavecontext] objects representing the stopped enclave will become unusable.

**Args**
* `enclaveIdentifier`: [Identifier][identifier] of the enclave to stop.

### `destroyEnclave(String enclaveIdentifier)`
Stops the enclave with the given [identifier][identifier] and destroys the enclave objects (containers, networks, etc.).

**NOTE:** Any [EnclaveContext][enclavecontext] objects representing the stopped enclave will become unusable.

**Args**
* `enclaveIdentifier`: [Identifier][identifier] of the enclave to destroy.

### `clean(boolean shouldCleanAll) -> []EnclaveNameAndUuid RemovedEnclaveNameAndUuids`
Destroys enclaves in the Kurtosis engine.

**Args**
* `shouldCleanAll`: If set to true, destroys running enclaves in addition to stopped ones.

**Returns**
* `RemovedEnclaveNameAndUuids`: A list of enclave uuids and names that were removed successfully

### `getServiceLogs(String enclaveIdentifier, Set<ServiceUUID> serviceUuids, Boolean shouldFollowLogs, LogLineFilter logLineFilter) -> ServiceLogsStreamContent serviceLogsStreamContent`
Get and start a service container logs stream (showed in ascending order, with the oldest line first) from services identified by their UUID.

**Args**
* `enclaveIdentifier`: [Identifier][identifier] of the services' enclave.
* `serviceUuids`: A set of service UUIDs identifying the services from which logs should be retrieved.
* `shouldFollowLogs`: If it's true, the stream will constantly send the new log lines. if it's false, the stream will be closed after the last created log line is sent.
* `logLineFilter`: The [filter][loglinefilter] that will be used for filtering the returned log lines

**Returns**
* `serviceLogsStreamContent`: The [ServiceLogsStreamContent][servicelogsstreamcontent] object which wrap all the information coming from the logs stream.

### `getExistingAndHistoricalEnclaveIdentifiers() -> EnclaveIdentifiers enclaveIdentifiers`

Get all (active & deleted) historical [identifiers][identifier] for the currently
running Kurtosis engine.

**Returns**
* `enclaveIdentifiers` The [EnclaveIdentifiers][enclave-identifiers] which provides user-friendly ways to lookup enclave identifier information.

EnclaveIdentifiers
-------------------
This class is a representation of identifiers of enclaves.

### `getEnclaveUuidForIdentifier(string identifier) -> EnclaveUUID enclaveUuid, Error`
Returns the UUID that matches the given identifier. If there are no matches it returns
an error instead.

**Args**
* `identifier`: A enclave identifier string

**Returns**
* `enclaveUuid`: The UUID for the enclave identified by the `identifier`.

### `getOrderedListOfNames() -> []String enclaveNames`
Returns an ordered list of names for all the enclaves registered with the engine. This is useful
if users want to enumerate all enclave names, say for an autocomplete like function.

**Returns**
* `enclaveNames`: This is a sorted list of enclave names

ServiceLogsStreamContent
------------------------
This class is the representation of the content sent during a service logs stream communication. This wrapper includes the service's logs content and the not found service UUIDs.

### `getServiceLogsByServiceUuids() ->  Map<ServiceUUID, Array<ServiceLog>> serviceLogsByServiceUuids`
Returns the user service logs content grouped by the service's UUID.

**Returns**
* `serviceLogsByServiceUuids`: A map containing a list of the [ServiceLog][servicelog] objects grouped by service UUID.

### `getNotFoundServiceUuids() -> Set<ServiceUUID> notFoundServiceUuids`
Returns the not found service UUIDs. The UUIDs may not be found either because they don't exist, or because the services haven't sent any logs.

**Returns**
* `notFoundServiceUuids`: A set of not found service UUIDs

ServiceLog
----------
This class represents single service's log line information

### `getContent() -> String content`

**Returns**
* `content`: The log line string content

LogLineFilter
-------------
This class is used to specify the match used for filtering the service's log lines. There are a couple of helpful constructors that can be used to generate the filter type

### `NewDoesContainTextLogLineFilter(String text) -> LogLineFilter logLineFilter`
Returns a LogLineFilter type which must be used for filtering the log lines containing the text match

**Args**
* `text`: The text that will be used to match in the log lines

**Returns**
* `logLineFilter`: The does-contain-text-match log line filter

### `NewDoesNotContainTextLogLineFilter(String text) -> LogLineFilter logLineFilter`
Returns a LogLineFilter type which must be used for filtering the log lines that do not contain the text match

**Args**
* `text`: The text that will be used to match in the log lines

**Returns**
* `logLineFilter`: The does-not-contain-text-match log line filter

### `NewDoesContainMatchRegexLogLineFilter(String regex) -> LogLineFilter logLineFilter`
Returns a LogLineFilter type which must be used for filtering the log lines containing the regex match, [re2 syntax regex may be used][google_re2_syntax_docs]

**Args**
* `regex`: The regex expression that will be used to match in the log lines

**Returns**
* `logLineFilter`: The does-contain-regex-match log line filter

### `NewDoesNotContainMatchRegexLogLineFilter(String regex) -> LogLineFilter logLineFilter`
Returns a LogLineFilter type which must be used for filtering the log lines that do not contain the regex match, [re2 syntax regex may be used][google_re2_syntax_docs]

**Args**
* `regex`: The regex expression that will be used to match in the log lines

**Returns**
* `logLineFilter`: The does-not-contain-regex-match log line filter

Enclaves
--------

This Kurtosis provided class is a collection of various different [EnclaveInfo][enclaveinfo] objects, by UUID, shortened UUID, and name.

### Map<String, EnclaveInfo> `enclavesByUuid`

A map from UUIDs to the enclave info for the enclave with the given UUID.

### Map<String, EnclaveInfo> `enclavesByName`

A map from names to the enclave info for the enclave with the given name

### Map<String, EnclaveInfo[]> `enclavesByShortenedUuid`

A map from shortened UUID (first 12 characters of UUID) to the enclave infos of the enclaves it matches too.

EnclaveInfo
-----------

This Kurtosis provided class contains information about enclaves. This class just contains data and no methods to manipulate enclaves. Users must use [EnclaveContext][enclavecontext] to modify the state of an enclave.

### `getEnclaveUuid() -> EnclaveUuid`
Gets the UUID of the enclave that this [EnclaveInfo][enclaveinfo] object represents.

### `getShortenedUuid() -> String`
Gets the shortened UUID of the enclave that this [EnclaveInfo][enclaveinfo] object represents.

### `getName() -> String`
Gets the name of the enclave that this [EnclaveInfo][enclaveinfo] object represents.

### `getCreationTime() -> Timestamp`
Gets the timestamp at which the enclave that this [EnclaveInfo][enclaveinfo] object represents was created.

### `getCreationTime() -> Timestamp`
Gets the timestamp at which the enclave that this [EnclaveInfo][enclaveinfo] object represents was created.

### `getContainersStatus() -> Status`
Gets the current status of the container running the enclave represented by this [EnclaveInfo][enclaveinfo]. Is one of 'EMPTY', 'RUNNING' and 'STOPPED'.

EnclaveContext
--------------
This Kurtosis-provided class is the lowest-level representation of a Kurtosis enclave, and provides methods for inspecting and manipulating the contents of the enclave. 

### `getEnclaveUuid() -> EnclaveUuid`
Gets the UUID of the enclave that this [EnclaveContext][enclavecontext] object represents.

### `getEnclaveName() -> String`
Gets the name of the enclave that this [EnclaveContext][enclavecontext] object represents.

### `runStarlarkScript(String mainFunctionName, String serializedStarlarkScript, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a provided Starlark script inside the enclave.

**Args**

* `mainFunctionName`: The main function name, an empty string can be passed to use the default value 'run'
* `serializedStarlarkScript`: The Starlark script provided as a string
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkPackage(String packageRootPath, String relativePathToMainFile, String mainFunctionName, String serializedParams, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a provided Starlark script inside the enclave.

**Args**

* `packageRootPath`: The path to the root of the package
* `relativePathToMainFile`: The relative filepath (relative to the package's root) to the main file, and empty string can be passed to use the default value 'main.star'. Example: if your main file is located in a path like this `github.com/my-org/my-package/src/internal/my-file.star` you should set `src/internal/my-file.star` as the relative path.
* `mainFunctionName`: The main function name, an empty string can be passed to use the default value 'run'.
* `serializedParams`: The parameters to pass to the package for the run. It should be a serialized JSON string.
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkRemotePackage(String packageId, String relativePathToMainFile, String mainFunctionName, String serializedParams, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a Starlark script hosted in a remote github.com repo inside the enclave.

**Args**

* `packageId`: The ID of the package pointing to the github.com repo hosting the package. For example `github.com/kurtosistech/datastore-army-package`
* `relativePathToMainFile`: The relative filepath (relative to the package's root) to the main file, and empty string can be passed to use the default value 'main.star'. Example: if your main file is located in a path like this `github.com/my-org/my-package/src/internal/my-file.star` you should set `src/internal/my-file.star` as the relative path.
* `mainFunctionName`: The main function name, an empty string can be passed to use the default value 'run'.
* `serializedParams`: The parameters to pass to the package for the run. It should be a serialized JSON string.
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkScriptBlocking(String mainFunctionName, String serializedStarlarkScript, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkScript][enclavecontext_runstarlarkscript], that blocks until the execution of the script is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

### `runStarlarkPackageBlocking(String packageRootPath, String relativePathToMainFile, String mainFunctionName, String serializedParams, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkPackage][enclavecontext_runstarlarkpackage], that blocks until the execution of the package is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

### `runStarlarkRemotePackageBlocking(String packageId, String relativePathToMainFile, String mainFunctionName, String serializedParams, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkRemotePackage][enclavecontext_runstarlarkremotepackage], that blocks until the execution of the package is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

### `getServiceContext(String serviceIdentifier) -> ServiceContext serviceContext`
Gets relevant information about a service (identified by the given service [identifier][identifier]) that is running in the enclave.

**Args**

* `serviceIdentifier`: The [identifier(name, UUID or short name)][identifier] of the target service

**Returns**

The [ServiceContext][servicecontext] representation of a service running in a Docker container.

### `getServices() -> Map<ServiceName,  ServiceUUID> serviceIdentifiers`
Gets the Name and UUID of the current services in the enclave.

**Returns**

* `serviceIdentifiers`: A map of objects containing a mapping of Name -> UUID for all the services inside the enclave

### `uploadFiles(String pathToUpload, String artifactName) -> FilesArtifactUUID, FilesArtifactName, Error`
Uploads a filepath or directory path as a [files artifact](./concepts-reference/files-artifacts.md). The resulting files artifact can be used in [`ServiceConfig.files`](./starlark-reference/service-config.md) when adding a service.

If a directory is specified, the contents of the directory will be uploaded to the archive without additional nesting. Empty directories cannot be uploaded.

**Args**

* `pathToUpload`: Filepath or dirpath on the local machine to compress and upload to Kurtosis.
* `artifactName`: The name to refer the artifact with.

**Returns**

* `FilesArtifactUUID`: A UUID identifying the new files artifact, which can be used in [`ServiceConfig.files`](./starlark-reference/service-config.md).
* `FilesArtifactName`: The name of the file-artifact, it is auto-generated if `artitfactName` is an empty string.

### `storeWebFiles(String urlToDownload, String artifactName)`
Downloads a files-containing `.tgz` from the given URL as a [files artifact](./concepts-reference/files-artifacts.md). The resulting files artifact can be used in [`ServiceConfig.files`](./starlark-reference/service-config.md) when adding a service.

**Args**

* `urlToDownload`: The URL on the web where the files-containing `.tgz` should be downloaded from.
* `artifactName`: The name to refer the artifact with.

**Returns**

* `UUID`: A UUID identifying the new files artifact, which can be used in [`ServiceConfig.files`](./starlark-reference/service-config.md).

### `getExistingAndHistoricalServiceIdentifiers() -> ServiceIdentifiers serviceIdentifiers`

Get all (active & deleted) historical [identifiers][identifier] for services for the enclave represented by the [EnclaveContext][enclavecontext].

**Returns**
* `serviceIdentifiers`: The [ServiceIdentifiers][service-identifiers] which provides user-friendly ways to lookup service identifier information.

### `getAllFilesArtifactNamesAndUuids() -> []FilesArtifactNameAndUuid filesArtifactNamesAndUuids`

Get a list of all files artifacts that are registered with the enclave represented by the [EnclaveContext][enclavecontext]

**Returns**
* `filesArtifactNameAndUuids`: A list of files artifact names and their corresponding uuids.

ServiceIdentifiers
-------------------
This class is a representation of service identifiers for a given enclave.

### `getServiceUuidForIdentifier(string identifier) -> ServiceUUID serviceUUID, Error`
Returns the UUID that matches the given identifier. If there are no matches it returns
an error instead.

**Args**
* `identifier`: A service identifier string

**Returns**
* `enclaveUuid`: The UUID for the service identified by the `identifier`.

### `getOrderedListOfNames() -> []String serviceNames`
Returns an ordered list of names for all the services in the enclave. This is useful
if users want to enumerate all service names, say for an autocomplete like function.

**Returns**
* `serviceNames`: This is a sorted list of service names

StarlarkRunResponseLine
-----------------------

This is a union object representing a single line returned by Kurtosis' Starlark runner. All Starlark run endpoints will return a stream of this object.

Each line is one of:

### [StarlarkInstruction][starlarkinstruction] `instruction`
An instruction that is _about to be_ executed. 

### [StarlarkInstructionResult][starlarkinstructionresult] `instructionResult`
The result of an instruction that was successfully executed

### [StarlarkError][starlarkerror] `error`
The error that was thrown running the Starlark code

### [StarlarkRunProgress][starlarkrunprogress] `progressInfo`
Regularly during the run of the code, Kurtosis' Starlark engine will send progress information through the stream to account for progress that was made running the code.

StarlarkInstruction
-------------------

`StarlarkInstruction` represents a Starlark instruction that is currently being executed. It contains the following fields:

* `instructionName`: the name of the instruction

* `instructionPosition`: the position of the instruction in the source code. It iscomposed of (filename, line number, column number)

* `arguments`: The list of arguments provided to this instruction. Each argument is composed of an optional name (if it was named in the source script) and its serialized value

* `executableInstruction`: A single string representing the instruction in valid Starlark code

StarlarkInstructionResult
-------------------------

`StarlarkInstructionResult` is the result of an instruction that was successfully run against Kurtosis engine. It is a single string field corresponding to the output of the instruction.

StarlarkError
-------------

Errors can be of three kind:

* Interpretation error: these errors happen before Kurtosis was able to execute the script. It typically means there's a syntax error in the provided Starlark code. The error message should point the users to where the code is incorrect.

* Validation error: these errors happen after interpretation was successful, but before the execution actually started in Kurtosis. Before starting the execution, Kurtosis runs some validation on the instructions that are about to be executed. The error message should contain more information on which instruction is incorrect.

* Execution error: these errors happen during the execution of the script against Kurtosis engine. More information is available in the error message.

StarlarkRunProgress
-------------------

`StarlarkRunProgress` accounts for progress that is made during a Starlark run. It contains three fields:

* `totalSteps`: The total number of steps for this run

* `currentStepNumber`: The number of the step that is currently being executed

* `currentStepInfo`: A string field with some information on the current step being executed.

StarlarkRunResult
-----------------

`StarlarkRunResult` is the object returned by the blocking functions to run Starlark code. It is similar to [StarlarkRunResponseLine][starlarkrunresponseline] except that it is not a union object:

* `instructions`: the [Starlark Instruction][starlarkinstruction] that were run

* `insterpretationError`: a potential Starlark Interpretation error (see [StarlarkError][starlarkerror]

* `validationErrors`: potential Starlark Validation errors (see [StarlarkError][starlarkerror]

* `executionError`: a potential Starlark Execution error (see [StarlarkError][starlarkerror]

* `runOutput`: The full output of the run, composed of the concatenated output for each instruction that was executed (separated by a newline)

ServiceContext
--------------
This Kurtosis-provided class is the lowest-level representation of a service running inside a Docker container. It is your handle for retrieving container information and manipulating the container.

### `getServiceName() -> ServiceName`
Gets the Name that Kurtosis uses to identify the service.

**Returns**

The service's Name.

### `getServiceUuid() -> ServiceUUID`
Gets the UUID (Universally Unique Identifier) that Kurtosis creates and uses to identify the service.
The differences with the Name is that this one is created by Kurtosis, users can't specify it, and this never can be repeated, every new execution of the same service will have a new UUID

**Returns**

The service's UUID.

### `getPrivateIpAddress() -> String`
Gets the IP address where the service is reachable at from _inside_ the enclave that the container is running inside. This IP address is how other containers inside the enclave can connect to the service.

**Returns**

The service's private IP address.

### `getPrivatePorts() -> Map<PortID, PortSpec>`
Gets the ports that the service is reachable at from _inside_ the enclave that the container is running inside. These ports are how other containers inside the enclave can connect to the service.

**Returns**

The ports that the service is reachable at from inside the enclave, identified by the user-chosen port ID set in [`ServiceConfig.ports`](./starlark-reference/service-config.md) when the service was created.

### `getMaybePublicIpAddress() -> String`
If the service declared used ports in [`ServiceConfig.ports`](./starlark-reference/service-config.md), then this function returns the IP address where the service is reachable at from _outside_ the enclave that the container is running inside. This IP address is how clients on the host machine can connect to the service. If no used ports were declared, this will be empty.

**Returns**

The service's public IP address, or an empty value if the service didn't declare any used ports.

### `getPublicPorts() -> Map<PortID, PortSpec>`
Gets the ports that the service is reachable at from _outside_ the enclave that the container is running inside. These ports are how clients on the host machine can connect to the service. If the service didn't declare any used ports in [`ServiceConfig.ports`](./starlark-reference/service-config.md), this value will be an empty map.

**Returns**

The ports (if any) that the service is reachable at from outside the enclave, identified by the user-chosen ID set in [`ServiceConfig.ports`](./starlark-reference/service-config.md) when the service was created.

### `execCommand(List<String> command) -> (int exitCode, String logs)`
Uses [Docker exec](https://docs.docker.com/engine/reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**

* `command`: The args of the command to execute in the container.

**Returns**

* `exitCode`: The exit code of the command.
* `logs`: The output of the run command, assuming a UTF-8 encoding. **NOTE:** Commands that output non-UTF-8 output will likely be garbled!

<!-------------------------------- ONLY LINKS BELOW HERE ------------------------>

<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[kurtosis-client-libs]: https://github.com/kurtosis-tech/kurtosis/tree/main/api

[servicelogsstreamcontent]: #servicelogsstreamcontent
[servicelog]: #servicelog

[enclavecontext]: #enclavecontext
[enclavecontext_runstarlarkscript]: #runstarlarkscriptstring-mainfunctionname-string-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
[enclavecontext_runstarlarkpackage]: #runstarlarkscriptstring-mainfunctionname-string-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
[enclavecontext_runstarlarkremotepackage]: #runstarlarkscriptstring-mainfunctionname-string-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error

[starlarkrunresponseline]: #starlarkrunresponseline
[starlarkinstruction]: #starlarkinstruction
[starlarkinstructionresult]: #starlarkinstructionresult
[starlarkerror]: #starlarkerror
[starlarkrunprogress]: #starlarkrunprogress
[starlarkrunresult]: #starlarkrunresult

[servicecontext]: #servicecontext
[servicecontext_getpublicports]: #getpublicports---mapportid-portspec
  
[loglinefilter]: #loglinefilter
[google_re2_syntax_docs]: https://github.com/google/re2/wiki/Syntax

[enclaveinfo]: #enclaveinfo
[enclaves]: #enclaves

[identifier]: ./concepts-reference/resource-identifier.md
[enclave-identifiers]: #enclaveidentifiers
[service-identifiers]: #serviceidentifiers
