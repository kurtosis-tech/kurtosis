Kurtosis Client Documentation
=============================
This documentation describes how to interact with the Kurtosis API from within a testnet. It includes information about starting services, stopping services, repartitioning the network, etc. Note that any comments specific to a language implementation will be found in the code comments.

_Found a bug? File it on [the repo][issues]!_



ModuleContext
-------------
This Kurtosis-provided class is the lowest-level representation of a Kurtosis module - a Docker container with a connection to the Kurtosis engine that responds to commands.

### execute(String serializedParams) -\> String serializedResult
Some modules are considered executable, meaning they respond to an "execute" command. This function will send the execute command to the module with the given serialized args, returning the serialized result. The serialization format of args & response will depend on the module. If the module isn't executable (i.e. doesn't respond to an "execute" command) then an error will be thrown.

**Args**

* `serializedParams`: Serialized data containing args to the module's execute function. Consult the documentation for the module you're using to determine what this should contain.

**Returns**

* `serializedResult`: Serialized data containing the results of executing the module. Consult the documentation for the module you're using to determine what this will contain.



EnclaveContext
--------------
This Kurtosis-provided class is the lowest-level representation of a Kurtosis enclave, and provides methods for inspecting and manipulating the contents of the enclave. 

### getEnclaveId() -\> EnclaveID
Gets the ID of the enclave that this [EnclaveContext][enclavecontext] object represents.

### loadModule(String moduleId, String image, String serializedParams) -\> [ModuleContext][modulecontext] moduleContext
Starts a new Kurtosis module (configured using the serialized params) inside the enclave, which makes it available for use.

**Args**

* `moduleId`: The ID that the new module should receive (must not exist).
* `image`: The container image of the module to be loaded.
* `serializedParams`: Serialized parameter data that will be passed to the module as it starts, to control overall module behaviour.

**Returns**

* `moduleContext`: The [ModuleContext][modulecontext] representation of the running module container, which allows execution of the execute function (if it exists).

### unloadModule(String moduleId) 
Stops and removes a Kurtosis module from the enclave.

**Args**

* `moduleId`: The ID of the module to remove.

### getModuleContext(String moduleId) -\> [ModuleContext][modulecontext] moduleContext
Gets the [ModuleContext][modulecontext] associated with an already-running module container identified by the given ID.

**Args**

* `moduleId`: The ID of the module to retrieve the context for.

**Returns**

* `moduleContext`: The [ModuleContext][modulecontext] representation of the running module container, which allows execution of the module's execute function (if it exists).

### executeStartosisScript(String serializedStartosisScript) -\> ([ExecuteStartosisResponse][executestartosisresponse] executionResult, Error error)

Execute a provide Startosis script inside the enclave.

**Args**

* `serializedStartosisScript`: The Startosis script provided as a string

**Returns**

* `executionResult`: The result of the execution, as a [ExecuteStartosisResponse][executestartosisresponse] object

### executeStartosisModule(String moduleRootPath) -\> ([ExecuteStartosisResponse][executestartosisresponse] executionResult, Error error)

Execute a provide Startosis script inside the enclave.

**Args**

* `moduleRootPath`: The path to the root of the module

**Returns**

* `executionResult`: The result of the execution, as a [ExecuteStartosisResponse][executestartosisresponse] object

<!-- TODO DELETE THIS!!! -->
### registerFilesArtifacts(Map\<FilesArtifactID, String\> filesArtifactUrls)
Downloads the given files artifacts to the Kurtosis engine, associating them with the given IDs, so they can be mounted inside a service's filespace at creation time via [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

**Args**

* `filesArtifactUrls`: A map of files_artifact_id -> url, where the ID is how the artifact will be referenced in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints] and the URL is the URL on the web where the files artifact should be downloaded from.

### addServiceToPartition(ServiceID serviceId, PartitionID partitionId, [ContainerConfig][containerconfig] containerConfig) -\> ([ServiceContext][servicecontext] serviceContext)
Starts a new service in the enclave with the given service ID, inside the partition with the given ID, using the given container config.

**Args**

* `serviceId`: The ID that the new service should have.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the enclave hasn't been repartitioned and the default partition removed).
* `containerConfig`: A [ContainerConfig][containerconfig] object indicating how to configure the service.

**Returns**

* `serviceContext`: The [ServiceContext][servicecontext] representation of a service running in a Docker container. Port information can be found in `ServiceContext.GetPublicPorts()`. The port spec strings that the service declared (as defined in [ContainerConfig.usedPorts][containerconfig_usedports]), mapped to the port on the host machine where the port has been bound to. This allows you to make requests to a service running in Kurtosis by making requests to a port on your local machine. If a port was not bound to a host machine port, it will not be present in the map (and if no ports were bound to host machine ports, the map will be empty).

### addServicesToPartition(Map\<ServiceID, [ContainerConfig][containerconfig]\> containerConfigs, PartitionID partitionId) -\> (Map\<ServiceID, [ServiceContext][servicecontext]\> successfulServices, Map\<ServiceID, Error\> failedServices)
Start services in bulk in the enclave with the given service IDs, inside the partition with the given ID, using the given container config.

**Args**

* `containerConfigs`: A mapping of service IDs to start in the enclave to their `containerConfig` indicating how to configure the service.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the enclave hasn't been repartitioned and the default partition removed).

**Returns**

* `successfulServices`: A mapping of service IDs that were successfully started in the enclave to their respective [ServiceContext][servicecontext] representation. 
* `failedServices`: A mapping of service IDs to the errors the caused that prevented the services from being added successfully to the enclave. 

### addService(ServiceID serviceId,  [ContainerConfig][containerconfig] containerConfig) -\> ([ServiceContext][servicecontext] serviceContext)
Convenience wrapper around [EnclaveContext.addServiceToPartition][enclavecontext_addservicetopartition], that adds the service to the default partition. Note that if the enclave has been repartitioned and the default partition doesn't exist anymore, this method will fail.

### addServices(Map\<ServiceID, [ContainerConfig][containerconfig]\> containerConfigs) -\> (Map\<ServiceID, [ServiceContext][servicecontext]\> successfulServices, Map\<ServiceID, Error\> failedServices)
Convenience wrapper around [EnclaveContext.addServicesToPartition][enclavecontext_addservicestopartition], that adds the services to the default partition. Note that if the enclave has been repartitioned and the default partition doesn't exist anymore, this method will fail.

### getServiceContext(ServiceID serviceId) -\> [ServiceContext][servicecontext]
Gets relevant information about a service (identified by the given service ID) that is running in the enclave.

**Args**

* `serviceId`: The ID of the service to pull the information from.

**Returns**

The [ServiceContext][servicecontext] representation of a service running in a Docker container.

### removeService(ServiceID serviceId, uint64 containerStopTimeoutSeconds)
Stops the container with the given service ID and removes it from the enclave.

**Args**

* `serviceId`: The ID of the service to remove.
* `containerStopTimeoutSeconds`: The number of seconds to wait for the container to gracefully stop before hard-killing it.

### repartitionNetwork(Map\<PartitionID, Set\<ServiceID\>\> partitionServices, Map\<PartitionID, Map\<PartitionID, [PartitionConnection][partitionconnection]\>\> partitionConnections, [PartitionConnection][partitionconnection] defaultConnection)
Repartitions the enclave so that the connections between services match the specified new state. All services currently in the enclave must be allocated to a new partition. 

**NOTE: For this to work, partitioning must be turned on when the Enclave is created with [KurtosisContext.createEnclave()][kurtosiscontext_createenclave].**

**Args**

* `partitionServices`: A definition of the new partitions in the enclave, and the services allocated to each partition. A service can only be allocated to a single partition.
* `partitionConnections`: Definitions of the connection state between the new partitions. If a connection between two partitions isn't defined in this map, the default connection will be used. Connections are not directional, so an error will be thrown if the same connection is defined twice (e.g. `Map[A][B] = someConnection`, and `Map[B][A] = otherConnection`).
* `defaultConnection`: The network state between two partitions that will be used if the connection isn't defined in the partition connections map.

### waitForHttpGetEndpointAvailability(ServiceID serviceId, uint32 port, String path, String requestBody, uint32 initialDelayMilliseconds, uint32 retries, uint32 retriesDelayMilliseconds, String bodyText)
Waits until a service endpoint is available by making requests to the endpoint using the given parameters, and the HTTP GET method. An error is thrown if the number of retries is exceeded.

**Args**

* `serviceId`: The ID of the service to check.
* `port`: The port (e.g. 8080) of the endpoint to check.
* `path`: The path of the service to check, which must not start with a slash (e.g. `service/health`).
* `initialDelayMilliseconds`: Number of milliseconds to wait until executing the first HTTP call
* `retries`: Max number of HTTP call attempts that this will execute until giving up and returning an error
* `retriesDelayMilliseconds`: Number of milliseconds to wait between retries
* `bodyText`: If this value is non-empty, the endpoint will not be marked as available until this value is returned (e.g. `Hello World`). If this value is emptystring, no body text comparison will be done.

### waitForHttpPostEndpointAvailability(ServiceID serviceId, uint32 port, String path, String requestBody, uint32 initialDelayMilliseconds, uint32 retries, uint32 retriesDelayMilliseconds, String bodyText)
Waits until a service endpoint is available by making requests to the endpoint using the given parameters, and the HTTP POST method. An error is thrown if the number of retries is exceeded.

**Args**

* `serviceId`: The ID of the service to check.
* `port`: The port (e.g. 8080) of the endpoint to check.
* `path`: The path of the service to check, which must not start with a slash (e.g. `service/health`).
* `requestBody`: The request body content that will be sent to the endpoint being checked.
* `initialDelayMilliseconds`: Number of milliseconds to wait until executing the first HTTP call
* `retries`: Max number of HTTP call attempts that this will execute until giving up and returning an error
* `retriesDelayMilliseconds`: Number of milliseconds to wait between retries
* `bodyText`: If this value is non-empty, the endpoint will not be marked as available until this value is returned (e.g. `Hello World`). If this value is emptystring, no body text comparison will be done.

### getServices() -\> Set\<ServiceID\> serviceIDs
Gets the IDs of the current services in the enclave.

**Returns**

* `serviceIDs`: A set of service IDs

### getModules() -\> Set\<ModuleID\> moduleIds
Gets the IDs of the Kurtosis modules that have been loaded into the enclave.

**Returns**

* `moduleIds`: A set of Kurtosis module IDs that are running in the enclave

### uploadFiles(String pathToUpload)
Takes a filepath or directory path that will be compressed and uploaded to the Kurtosis filestore for use with [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

If a directory is specified, the _contents_ of the directory will be uploaded to the archive without additional nesting.
Empty directories cannot be uploaded.


**Args**

* `pathToUpload`: Filepath or dirpath on the local machine to compress and upload to Kurtosis.

**Returns**

* `uuid`: A unique ID as a string identifying the uploaded files, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

### storeWebFiles(String urlToDownload)
Downloads a files-containing `.tgz` from the given URL to the Kurtosis engine, so that the files inside can be mounted inside a service's filespace at creation time via [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

**Args**

* `urlToDownload`: The URL on the web where the files-containing `.tgz` should be downloaded from.

**Returns**

* `uuid`: A unique ID as a string identifying the downloaded, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

### storeServiceFiles(ServiceID serviceId, String absoluteFilepathOnServiceContainer)
Copy a file or folder from a service container to the Kurtosis filestore for use with [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints]

**Args**

* `serviceId`: The ID of the service which contains the file or the folder.
* `absoluteFilepathOnServiceContainer`: The absolute filepath on the service where the file or folder should be copied from

**Returns**

* `uuid`: A unique ID as a string identifying the generated files artifact, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

### pauseService(ServiceID serviceId)
Pauses all running processes in the specified service, but does not shut down the service. Processes can be restarted with [EnclaveContext.unpauseService][enclavecontext_unpauseservice].

**Args**

* `serviceId`: The ID of the service to pause.

### unpauseService(ServiceID serviceId)
Unpauses all paused processes in the specified service. Specified service must have been previously paused.

**Args**

* `serviceId`: The ID of the service to unpause.

### renderTemplates(Map<String, TemplateAndData> templateAndDataByDestinationRelFilepaths)
Renders templates and stores them in an archive that gets uploaded to the Kurtosis filestore for use with [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

The destination relative filepaths are relative to the root of the archive that gets stored in the filestore.

**Args**

* `templateAndDataByDestinationRelFilepaths`: A map of [TemplateAndData](#templateanddata) by the destination relative file path.

**Returns**

* `uuid`: A unique ID as a string identifying the archived rendered templates, which can be used in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

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

### Map\<PortID, PortSpec\> usedPorts
The ports that the container will be listening on, identified by a user-friendly ID that can be used to select the port again in the future (e.g. via [ServiceContext.getPublicPorts][servicecontext_getpublicports].

### Map\<String, String\> filesArtifactMountpoints
Sometimes a service needs files to be available before it starts (e.g. starting a service with a 5 GB Postgres database mounted). To ease this pain, Kurtosis allows you to specify gzipped TAR files that Kurtosis will uncompress and mount at locations on your service containers. These "files artifacts" will need to have been stored in Kurtosis beforehand using methods like [EnclaveContext.uploadFiles][enclavecontext_uploadfiles].

This property is therefore a map of the files artifact ID -> path on the container where the uncompressed artifact contents should be mounted, with the file artifact IDs corresponding to the ID returned by files-storing methods like [EnclaveContext.uploadFiles][enclavecontext_uploadfiles]. 

E.g. if I've previously uploaded a set of files using [EnclaveContext.uploadFiles][enclavecontext_uploadfiles] and Kurtosis has returned me the ID `813bdb20-3aab-4c5b-a0f5-a7deba7bf0d7`, I might ask Kurtosis to mount the contents inside my container at the `/database` path using a map like `{"813bdb20-3aab-4c5b-a0f5-a7deba7bf0d7": "/database"}`.

### List\<String\> entrypointOverrideArgs
You often won't control the container images that you'll be using in your testnet, and the `ENTRYPOINT` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### List\<String\> cmdOverrideArgs
You often won't control the container images that you'll be using in your testnet, and the `CMD` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### Map\<String, String\> environmentVariableOverrides
Defines environment variables that should be set inside the Docker container running the service. This can be necessary for starting containers from Docker images you don't control, as they'll often be parameterized with environment variables.

### uint64 cpuAllocationMillicpus
Allows you to set an allocation for CPU resources available in the underlying host container of a service. The metric used to measure `cpuAllocation`  is `millicpus`, 1000 millicpus is equivalent to 1 CPU on the underlying machine. This metric is identical [Docker's measure of `cpus`](https://docs.docker.com/config/containers/resource_constraints/#:~:text=Description-,%2D%2Dcpus%3D%3Cvalue%3E,-Specify%20how%20much) and [Kubernetes measure of `cpus` for limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu). Setting `cpuAllocationMillicpus=1500` is equivalent to setting `cpus=1.5` in Docker and `cpus=1.5` or `cpus=1500m` in Kubernetes. If set, the value must be a nonzero positive integer. If unset, there will be no constraints on CPU usage of the host container. 

### uint64 memoryAllocationMegabytes
Allows you to set an allocation for memory resources available in the underlying host container of a service. The metric used to measure `memoryAllocation` is `megabytes`. Setting `memoryAllocation=1000` is equivalent to setting the memory limit of the underlying host machine to `1e9 bytes` or `1GB`. If set, the value must be a nonzero positive integer of at least `6 megabytes` as Docker requires this as a minimum. If unset, there will be no constraints on memory usage of the host container. For information on memory limits in your underlying container engine, view [Docker](https://docs.docker.com/config/containers/resource_constraints/)'s and [Kubernetes](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)` docs.

### String privateIPAddrPlaceholder
The placeholder string used within `entrypointOverrideArgs`, `cmdOverrideArgs`, and `environmentVariableOverrides` that gets replaced with the private IP address of the container inside Docker/Kubernetes before the container starts. This defaults to `KURTOSIS_IP_ADDR_PLACEHOLDER` if this isn't set.
The user needs to make sure that they provide the same placeholder string for this field that they use in `entrypointOverrideArgs`, `cmdOverrideArgs`, and `environmentVariableOverrides`.


ContainerConfigBuilder
------------------------------
The builder that should be used to create [ContainerConfig][containerconfig] instances. The functions on this builder will correspond to the properties on the [ContainerConfig][containerconfig] object, in the form `withPropertyName` (e.g. `withUsedPorts` sets the ports used by the container).


ExecuteStartosisResponse
----------------------------

This is an object representing the result of the execution of a Startosis script. It has 4 fields:

### String `interpretationError`
An interpretation error is returned if the script couldn't be interpreted by Kurtosis backend

### String `validationError`
A validationError is returned if the script was successfully interpreted but could not be validated by Kurtosis backend

### String `executionError`
An execution error is returned if the script failed during its execution by Kurtosis backend

### String `scriptOutput`
The output of the script that was successfully executed inside the enclave


ServiceContext
--------------
This Kurtosis-provided class is the lowest-level representation of a service running inside a Docker container. It is your handle for retrieving container information and manipulating the container.

### getServiceId() -\> ServiceID
Gets the ID that Kurtosis uses to identify the service.

**Returns**

The service's ID.

### getPrivateIpAddress() -\> String
Gets the IP address where the service is reachable at from _inside_ the enclave that the container is running inside. This IP address is how other containers inside the enclave can connect to the service.

**Returns**

The service's private IP address.

### getPrivatePorts() -\> Map\<PortID, PortSpec\>
Gets the ports that the service is reachable at from _inside_ the enclave that the container is running inside. These ports are how other containers inside the enclave can connect to the service.

**Returns**

The ports that the service is reachable at from inside the enclave, identified by the user-chosen ID set in [ContainerConfig.usedPorts][containerconfig_usedports] when the service was created.

### getMaybePublicIpAddress() -\> String
If the service declared used ports in [ContainerConfig.usedPorts][containerconfig_usedports], then this function returns the IP address where the service is reachable at from _outside_ the enclave that the container is running inside. This IP address is how clients on the host machine can connect to the service. If no used ports were declared, this will be empty.

**Returns**

The service's public IP address, or an empty value if the service didn't declare any used ports.

### getPublicPorts() -\> Map\<PortID, PortSpec\>
Gets the ports that the service is reachable at from _outside_ the enclave that the container is running inside. These ports are how clients on the host machine can connect to the service. If the service didn't declare any used ports in [ContainerConfig.usedPorts][containerconfig_usedports], this value will be an empty map.

**Returns**

The ports (if any) that the service is reachable at from outside the enclave, identified by the user-chosen ID set in [ContainerConfig.usedPorts][containerconfig_usedports] when the service was created.

### execCommand(List\<String\> command) -\> (int exitCode, String logs)
Uses [Docker exec](https://docs.docker.com/engine/reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**

* `command`: The args of the command to execute in the container.

**Returns**

* `exitCode`: The exit code of the command.
* `logs`: The output of the run command, assuming a UTF-8 encoding. **NOTE:** Commands that output non-UTF-8 output will likely be garbled!




TemplateAndData
------------------

This is an object that gets used by the [renderTemplates](#rendertemplatesmapstring-templateanddata-templateanddatabydestinationrelfilepaths) function.
It has two properties.

### String template
The template that needs to be rendered. We support Golang [templates](https://pkg.go.dev/text/template). The casing of the keys inside the template and data doesn't matter.
### Any templateData
The data that needs to be rendered in the template. This will be converted into a JSON string before it gets sent over the wire. The elements inside the object should exactly match the keys in the template.

Note, because of how we handle floating point numbers & large integers, if you pass a floating point number it will get
printed in the decimal notation by default. If you want to use modifiers like `{{printf .%2f | .MyFloat}}`, you'll have to use
the `Float64` method on the `json.Number` first, so above would look like `{{printf .2%f | .MyFloat.Float64}}`.

---

_Found a bug? File it on [the repo][issues]!_

[issues]: https://github.com/kurtosis-tech/kurtosis-sdk/issues


<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[containerconfig]: #containerconfig
[containerconfig_usedports]: #mapportid-portspec-usedports
[containerconfig_filesartifactmountpoints]: #mapstring-string-filesartifactmountpoints

[containerconfigbuilder]: #containerconfigbuilder

[modulecontext]: #modulecontext

[enclavecontext]: #enclavecontext
[enclavecontext_registerfilesartifacts]: #registerfilesartifactsmapfilesartifactid-string-filesartifacturls
[enclavecontext_addservice]: #addserviceserviceid-serviceid--containerconfig-containerconfig---servicecontext-servicecontext
[enclavecontext_addservicetopartition]: #addservicetopartitionserviceid-serviceid-partitionid-partitionid-containerconfig-containerconfig---servicecontext-servicecontext
[enclavecontext_addservicestopartition]: #addservicestopartitionmapserviceid-containerconfig-containerconfigs-partitionid-partitionid---mapserviceid-servicecontext-successfulservices-mapserviceid-error-failedservices
[enclavecontext_unpauseservice]: #unpauseserviceserviceid-serviceid
[enclavecontext_repartitionnetwork]: #repartitionnetworkmappartitionid-setserviceid-partitionservices-mappartitionid-mappartitionid-partitionconnectionpartitionconnection-partitionconnections-partitionconnectionpartitionconnection-defaultconnection
[enclavecontext_uploadfiles]: #uploadfilesstring-pathtoupload
[enclavecontext_rendertemplates]: #rendertemplatesmapstring-templateanddata-templateanddatabydestinationrelfilepaths

[partitionconnection]: #partitionconnection

[executestartosisresponse]: #executestartosisresponse

[servicecontext]: #servicecontext
[servicecontext_getpublicports]: #getpublicports---mapportid-portspec

[templateanddata]: #templateanddata

[kurtosiscontext_createenclave]: ./engine-lib-documentation.md#createenclaveenclaveid-enclaveid-boolean-ispartitioningenabled---enclavecontextenclavecontext-enclavecontext
