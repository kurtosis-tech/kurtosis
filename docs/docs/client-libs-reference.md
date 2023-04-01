---
title: Client Libraries
sidebar_label: Client Libraries Reference
slug: /client-libs-reference
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

### `createEnclave(String enclaveName, boolean isPartitioningEnabled) -> [EnclaveContext][enclavecontext] enclaveContext`
Creates a new Kurtosis enclave using the given parameters.

**Args**
* `enclaveName`: The name to give the new enclave.
* `isPartitioningEnabled`: If set to true, the enclave will be set up to allow for repartitioning. This will make service addition & removal take slightly longer, but allow for calls to [EnclaveContext.repartitionNetwork][enclavecontext_repartitionnetwork].

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
* `RemovedEnclaveNameAndUuids`: A list of enclave uuids and names that were removed succesfully

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
This class represent single service's log line information

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

### `runStarlarkScript(String serializedStarlarkScript, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a provided Starlark script inside the enclave.

**Args**

* `serializedStarlarkScript`: The Starlark script provided as a string
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkPackage(String packageRootPath, String serializedParams, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a provided Starlark script inside the enclave.

**Args**

* `packageRootPath`: The path to the root of the package
* `serializedParams`: The parameters to pass to the package for the run. It should be a serialized JSON string.
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkRemotePackage(String packageId, String serializedParams, Boolean dryRun) -> (Stream<StarlarkRunResponseLine> responseLines, Error error)`

Run a Starlark script hosted in a remote github.com repo inside the enclave.

**Args**

* `packageId`: The ID of the package pointing to the github.com repo hosting the package. For example `github.com/kurtosistech/datastore-army-package`
* `serializedParams`: The parameters to pass to the package for the run. It should be a serialized JSON string.
* `dryRun`: When set to true, the Kurtosis instructions are not executed.

**Returns**

* `responseLines`: A stream of [StarlarkRunResponseLine][starlarkrunresponseline] objects

### `runStarlarkScriptBlocking(String serializedStarlarkScript, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkScript][enclavecontext_runstarlarkscript], that blocks until the execution of the script is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

### `runStarlarkPackageBlocking(String packageRootPath, String serializedParams, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkPackage][enclavecontext_runstarlarkpackage], that blocks until the execution of the package is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

### `runStarlarkRemotePackageBlocking(String packageId, String serializedParams, Boolean dryRun) -> (StarlarkRunResult runResult, Error error)`

Convenience wrapper around [EnclaveContext.runStarlarkRemotePackage][enclavecontext_runstarlarkremotepackage], that blocks until the execution of the package is finished and returns a single [StarlarkRunResult][starlarkrunresult] object containing the result of the run.

<!-- TODO DELETE THIS!!! -->
### `registerFilesArtifacts(Map<FilesArtifactID, String> filesArtifactUrls)`
Downloads the given files artifacts to the Kurtosis engine, associating them with the given IDs, so they can be mounted inside a service's filespace at creation time via [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

**Args**

* `filesArtifactUrls`: A map of files_artifact_id -> url, where the ID is how the artifact will be referenced in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints] and the URL is the URL on the web where the files artifact should be downloaded from.

### `addServiceToPartition(ServiceID serviceId, PartitionID partitionId, ContainerConfig containerConfig) -> ServiceContext serviceContext`
Starts a new service in the enclave with the given service ID, inside the partition with the given ID, using the given container config.

**Args**

* `serviceId`: The ID that the new service should have.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the enclave hasn't been repartitioned and the default partition removed).
* `containerConfig`: A [ContainerConfig][containerconfig] object indicating how to configure the service.

**Returns**

* `serviceContext`: The [ServiceContext][servicecontext] representation of a service running in a Docker container. Port information can be found in `ServiceContext.GetPublicPorts()`. The port spec strings that the service declared (as defined in [ContainerConfig.usedPorts][containerconfig_usedports]), mapped to the port on the host machine where the port has been bound to. This allows you to make requests to a service running in Kurtosis by making requests to a port on your local machine. If a port was not bound to a host machine port, it will not be present in the map (and if no ports were bound to host machine ports, the map will be empty).

### `addServicesToPartition(Map<ServiceID, ContainerConfig> containerConfigs, PartitionID partitionId) -> (Map<ServiceID, ServiceContext> successfulServices, Map<ServiceID, Error> failedServices)`
Start services in bulk in the enclave with the given service IDs, inside the partition with the given ID, using the given container config.

**Args**

* `containerConfigs`: A mapping of service IDs to start in the enclave to their `containerConfig` indicating how to configure the service.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the enclave hasn't been repartitioned and the default partition removed).

**Returns**

* `successfulServices`: A mapping of service IDs that were successfully started in the enclave to their respective [ServiceContext][servicecontext] representation.
* `failedServices`: A mapping of service IDs to the errors the caused that prevented the services from being added successfully to the enclave.

### `addService(ServiceID serviceId,  ContainerConfig containerConfig) -> (ServiceContext serviceContext)`
Convenience wrapper around [EnclaveContext.addServiceToPartition][enclavecontext_addservicetopartition], that adds the service to the default partition. Note that if the enclave has been repartitioned and the default partition doesn't exist anymore, this method will fail.

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

### `uploadFiles(String pathToUpload, String artifactName) -> FileArtifactUUID, FileArtifactName, Error`
Takes a filepath or directory path that will be compressed and uploaded to the Kurtosis filestore for use with [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

If a directory is specified, the contents of the directory will be uploaded to the archive without additional nesting. Empty directories cannot be uploaded.

**Args**

* `pathToUpload`: Filepath or dirpath on the local machine to compress and upload to Kurtosis.
* `artifactName`: The name to refer the artifact with.

**Returns**

* `FileArtifactUUID`: A unique ID as a string identifying the uploaded files, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].
* `FileArtifactName`: The name of the file-artifact, it is auto-generated if `artitfactName` is an empty string.

### `storeWebFiles(String urlToDownload, String artifactName)`
Downloads a files-containing `.tgz` from the given URL to the Kurtosis engine, so that the files inside can be mounted inside a service's filespace at creation time via [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

**Args**

* `urlToDownload`: The URL on the web where the files-containing `.tgz` should be downloaded from.
* `artifactName`: The name to refer the artifact with.

**Returns**

* `UUID`: A unique ID as a string identifying the downloaded, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

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

ModuleContext
-------------
**DEPRECATED: Use `runStarlarkScript` and `runStarlarkPackage` instead**

This Kurtosis-provided class is the lowest-level representation of a Kurtosis module - a Docker container with a connection to the Kurtosis engine that responds to commands.

### `execute(String serializedParams) -> String serializedResult`
**DEPRECATED: Use `runStarlarkScript` and `runStarlarkPackage` instead**

Some modules are considered executable, meaning they respond to an "execute" command. This function will send the execute command to the module with the given serialized args, returning the serialized result. The serialization format of args & response will depend on the module. If the module isn't executable (i.e. doesn't respond to an "execute" command) then an error will be thrown.

**Args**

* `serializedParams`: Serialized data containing args to the module's execute function. Consult the documentation for the module you're using to determine what this should contain.

**Returns**

* `serializedResult`: Serialized data containing the results of executing the module. Consult the documentation for the module you're using to determine what this will contain.

PartitionConnection
-------------------
This interface represents the network state between two partitions (e.g. whether network traffic is blocked, being partially dropped, etc.).

The three types of partition connections are: unblocked (all traffic is allowed), blocked (no traffic is allowed), and soft (packets are partially dropped). Each type of partition connection has a constructor that can be used to create them.

The soft partition constructor receives one parameter, `packetLossPercentage`, which sets the percentage of packet loss in the connection between the services that are part of the partition.

Unblocked partitions and blocked partitions have parameter-less constructors.

ContainerConfig
---------------
Object containing information Kurtosis needs to create and run the container. This config should be created using [ContainerConfigBuilder][containerconfigbuilder] instances.

### String image
The name of the container image that Kurtosis should use when creating the service's container (e.g. `my-repo/my-image:some-tag-name`).

### `Map<PortID, PortSpec> usedPorts`
The ports that the container will be listening on, identified by a user-friendly ID that can be used to select the port again in the future (e.g. via [ServiceContext.getPublicPorts][servicecontext_getpublicports].

### `Map<String, String> filesArtifactMountpoints`
Sometimes a service needs files to be available before it starts (e.g. starting a service with a 5 GB Postgres database mounted). To ease this pain, Kurtosis allows you to specify gzipped TAR files that Kurtosis will uncompress and mount at locations on your service containers. These "files artifacts" will need to have been stored in Kurtosis beforehand using methods like [EnclaveContext.uploadFiles][enclavecontext_uploadfiles].

This property is therefore a map of the files artifact ID -> path on the container where the uncompressed artifact contents should be mounted, with the file artifact IDs corresponding to the ID returned by files-storing methods like [EnclaveContext.uploadFiles][enclavecontext_uploadfiles].

E.g. if I've previously uploaded a set of files using [EnclaveContext.uploadFiles][enclavecontext_uploadfiles] and Kurtosis has returned me the ID `813bdb20-3aab-4c5b-a0f5-a7deba7bf0d7`, I might ask Kurtosis to mount the contents inside my container at the `/database` path using a map like `{"813bdb20-3aab-4c5b-a0f5-a7deba7bf0d7": "/database"}`.

### `List<String> entrypointOverrideArgs`
You often won't control the container images that you'll be using in your testnet, and the `ENTRYPOINT` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### `List<String> cmdOverrideArgs`
You often won't control the container images that you'll be using in your testnet, and the `CMD` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### `Map<String, String> environmentVariableOverrides`
Defines environment variables that should be set inside the Docker container running the service. This can be necessary for starting containers from Docker images you don't control, as they'll often be parameterized with environment variables.

### `uint64 cpuAllocationMillicpus`
Allows you to set an allocation for CPU resources available in the underlying host container of a service. The metric used to measure `cpuAllocation`  is `millicpus`, 1000 millicpus is equivalent to 1 CPU on the underlying machine. This metric is identical [Docker's measure of `cpus`](https://docs.docker.com/config/containers/resource_constraints/#:~:text=Description-,%2D%2Dcpus%3D%3Cvalue%3E,-Specify%20how%20much) and [Kubernetes measure of `cpus` for limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu). Setting `cpuAllocationMillicpus=1500` is equivalent to setting `cpus=1.5` in Docker and `cpus=1.5` or `cpus=1500m` in Kubernetes. If set, the value must be a nonzero positive integer. If unset, there will be no constraints on CPU usage of the host container. 

### `uint64 memoryAllocationMegabytes`
Allows you to set an allocation for memory resources available in the underlying host container of a service. The metric used to measure `memoryAllocation` is `megabytes`. Setting `memoryAllocation=1000` is equivalent to setting the memory limit of the underlying host machine to `1e9 bytes` or `1GB`. If set, the value must be a nonzero positive integer of at least `6 megabytes` as Docker requires this as a minimum. If unset, there will be no constraints on memory usage of the host container. For information on memory limits in your underlying container engine, view [Docker](https://docs.docker.com/config/containers/resource_constraints/)'s and [Kubernetes](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)'s docs.

### `String privateIPAddrPlaceholder`
The placeholder string used within `entrypointOverrideArgs`, `cmdOverrideArgs`, and `environmentVariableOverrides` that gets replaced with the private IP address of the container inside Docker/Kubernetes before the container starts. This defaults to `KURTOSIS_IP_ADDR_PLACEHOLDER` if this isn't set.
The user needs to make sure that they provide the same placeholder string for this field that they use in `entrypointOverrideArgs`, `cmdOverrideArgs`, and `environmentVariableOverrides`.


ContainerConfigBuilder
------------------------------
The builder that should be used to create [ContainerConfig][containerconfig] instances. The functions on this builder will correspond to the properties on the [ContainerConfig][containerconfig] object, in the form `withPropertyName` (e.g. `withUsedPorts` sets the ports used by the container).


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

* Interpretation error: these errors happens before Kurtosis was able to execute the script. It typically means there's a syntax error in the provided Starlark code. The error message should point the users to where the code is incorrect.

* Validation error: these errors happens after interpretation was successful, but before the execution actually started in Kurtosis. Before starting the execution, Kurtosis runs some validation on the instructions that are about to be executed. The error message should contain more information on which instruction is incorrect.

* Execution error: these errors happens during the execution of the script against Kurtosis engine. More information is available in the error message.

StarlarkRunProgress
-------------------

`StarlarkRunProgress` accounts for progress that is made during a Starlark run. It contains three fields:

* `totalSteps`: The total number of steps for this run

* `currentStepNumber`: The number of the step that is currently being executed

* `currentStepInfo`: A string field with some information on the current step being executed.

StarlarkRunResult
-----------------

`StarlarkRunResult` is the object returned by the blocking functions to run Starlark code. It is similar to [RunStarlarkResponseLine][runstarlarkresponseline] except that it is not a union object:

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

The ports that the service is reachable at from inside the enclave, identified by the user-chosen ID set in [ContainerConfig.usedPorts][containerconfig_usedports] when the service was created.

### `getMaybePublicIpAddress() -> String`
If the service declared used ports in [ContainerConfig.usedPorts][containerconfig_usedports], then this function returns the IP address where the service is reachable at from _outside_ the enclave that the container is running inside. This IP address is how clients on the host machine can connect to the service. If no used ports were declared, this will be empty.

**Returns**

The service's public IP address, or an empty value if the service didn't declare any used ports.

### `getPublicPorts() -> Map<PortID, PortSpec>`
Gets the ports that the service is reachable at from _outside_ the enclave that the container is running inside. These ports are how clients on the host machine can connect to the service. If the service didn't declare any used ports in [ContainerConfig.usedPorts][containerconfig_usedports], this value will be an empty map.

**Returns**

The ports (if any) that the service is reachable at from outside the enclave, identified by the user-chosen ID set in [ContainerConfig.usedPorts][containerconfig_usedports] when the service was created.

### `execCommand(List<String> command) -> (int exitCode, String logs)`
Uses [Docker exec](https://docs.docker.com/engine/concepts-reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**

* `command`: The args of the command to execute in the container.

**Returns**

* `exitCode`: The exit code of the command.
* `logs`: The output of the run command, assuming a UTF-8 encoding. **NOTE:** Commands that output non-UTF-8 output will likely be garbled!

TemplateAndData
------------------

This is an object that gets used by the [renderTemplates][enclavecontext_rendertemplates] function.
It has two properties.

### String template
The template that needs to be rendered. We support Golang [templates](https://pkg.go.dev/text/template). The casing of the `keys` or `fields` inside the template must match the casing of the `fields` or the `keys` inside the data.

### Any templateData
The data that needs to be rendered in the template. This will be converted into a JSON string before it gets sent over the wire. The elements inside the object should exactly match the keys in the template. If you are using a struct for `templateData` then the field names must start with an upper case letter to ensure that the field names are accessible outside of the structs own package. If you are using a map then you can keys that begin with lower case letters as well.

Note, because of how we handle floating point numbers & large integers, if you pass a floating point number it will get
printed in the decimal notation by default. If you want to use modifiers like `{{printf .%2f | .MyFloat}}`, you'll have to use
the `Float64` method on the `json.Number` first, so above would look like `{{printf .2%f | .MyFloat.Float64}}`.

<!-------------------------------- ONLY LINKS BELOW HERE ------------------------>

<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[kurtosis-client-libs]: https://github.com/kurtosis-tech/kurtosis/tree/main/api

[servicelogsstreamcontent]: #servicelogsstreamcontent
[servicelog]: #servicelog

[containerconfig]: #containerconfig
[containerconfig_usedports]: #mapportid-portspec-usedports
[containerconfig_filesartifactmountpoints]: #mapstring-string-filesartifactmountpoints

[containerconfigbuilder]: #containerconfigbuilder

[modulecontext]: #modulecontext

[enclavecontext]: #enclavecontext
[enclavecontext_registerfilesartifacts]: #registerfilesartifactsmapfilesartifactid-string-filesartifacturls

[partitionconnection]: #partitionconnection

[enclavecontext_runstarlarkscript]: #runstarlarkscriptstring-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
[enclavecontext_runstarlarkpackage]: #runstarlarkpackagestring-packagerootpath-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
[enclavecontext_runstarlarkremotepackage]: #runstarlarkremotepackagestring-packageid-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error

[starlarkrunresponseline]: #starlarkrunresponseline
[starlarkinstruction]: #starlarkinstruction
[starlarkinstructionresult]: #starlarkinstructionresult
[starlarkerror]: #starlarkerror
[starlarkrunprogress]: #starlarkrunprogress

[servicecontext]: #servicecontext
[servicecontext_getpublicports]: #getpublicports---mapportid-portspec
  
[templateanddata]: #templateanddata

[loglinefilter]: #loglinefilter
[google_re2_syntax_docs]: https://github.com/google/re2/wiki/Syntax

[enclaveinfo]: #enclaveinfo
[enclaves]: #enclaves

[identifier]: ./concepts-reference/resource-identifier.md
[enclave-identifiers]: #enclaveidentifiers
[service-identifiers]: #serviceidentifiers
