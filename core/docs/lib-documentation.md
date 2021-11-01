Kurtosis Client Documentation
=============================
This documentation describes how to interact with the Kurtosis API from within a testnet. It includes information about starting services, stopping services, repartitioning the network, etc. These objects are heavily used inside the [Kurtosis testing framework](../kurtosis-testsuite-api-lib/lib-documentation). Note that any comments specific to a language implementation will be found in the code comments.

_Found a bug? File it on [the repo](https://github.com/kurtosis-tech/kurtosis-client/issues)!_



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

### registerFilesArtifacts(Map\<FilesArtifactID, String\> filesArtifactUrls)
Downloads the given files artifacts to the Kurtosis engine, associating them with the given IDs, so they can be mounted inside a service's filespace at creation time via [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints].

**Args**

* `filesArtifactUrls`: A map of files_artifact_id -> url, where the ID is how the artifact will be referenced in [ContainerConfig.filesArtifactMountpoints][containerconfig_filesartifactmountpoints] and the URL is the URL on the web where the files artifact should be downloaded from.

### addServiceToPartition(ServiceID serviceId, PartitionID partitionId, Func(String ipAddr, [SharedPath][sharedpath] sharedDirectory) -\> [ContainerConfig][containerconfig] containerConfigSupplier) -\> ([ServiceContext][servicecontext] serviceContext, Map\<String, PortBinding\> hostPortBindings)
Starts a new service in the enclave with the given service ID, inside the partition with the given ID, using the given config supplier.

**Args**

* `serviceId`: The ID that the new service should have.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the enclave hasn't been repartitioned and the default partition removed).
* `containerConfigSupplier`: An anonymous function, used to produce the [ContainerConfig][containerconfig] for starting the service, which receives two dynamic values as arguments: 
    1. The IP address of the service being started
    1. A [SharedPath][sharedpath] object which represents a shared directory that is mounted on both a) the container where this code is running and b) the service container being started, so that files can be made available to the service container by creating them with this container. E.g. calling `sharedDirectory.getChildPath("newfile.txt")` will get the path to an object that can be a) written by this container via [SharedPath.getAbsPathOnThisContainer][sharedpath_getabspathonthiscontainer] and b) used by the service container via [SharedPath.getAbsPathOnServiceContainer][sharedpath_getabspathonservicecontainer].


**Returns**

* `serviceContext`: The [ServiceContext][servicecontext] representation of a service running in a Docker container.
* `hostPortBindings`: The port spec strings that the service declared (as defined in [ContainerConfig.usedPorts][containerconfig_usedports]), mapped to the port on the host machine where the port has been bound to. This allows you to make requests to a service running in Kurtosis by making requests to a port on your local machine. If a port was not bound to a host machine port, it will not be present in the map (and if no ports were bound to host machine ports, the map will be empty).

### addService(ServiceID serviceId,  Func(String ipAddr, [SharedPath][sharedpath] sharedDirectory) -\> [ContainerConfig][containerconfig] containerConfigSupplier) -\> ([ServiceContext][servicecontext] serviceContext, Map\<String, PortBinding\> hostPortBindings)
Convenience wrapper around [EnclaveContext.addServiceToPartition][enclavecontext_addservicetopartition], that adds the service to the default partition. Note that if the enclave has been repartitioned and the default partition doesn't exist anymore, this method will fail.

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

### repartitionNetwork(Map\<PartitionID, Set\<ServiceID\>> partitionServices, Map\<PartitionID, Map\<PartitionID, [PartitionConnectionInfo][partitionconnectioninfo]\>> partitionConnections, [PartitionConnectionInfo][partitionconnectioninfo] defaultConnection)
Repartitions the enclave so that the connections between services match the specified new state. All services currently in the enclave must be allocated to a new partition. 

**NOTE: For this to work, partitioning must be turned on in the [Test.configure][test_configure] method.**

**Args**

* `partitionServices`: A definition of the new partitions in the enclave, and the services allocated to each partition. A service can only be allocated to a single partition.
* `partitionConnections`: Definitions of the connection state between the new partitions. If a connection between two partitions isn't defined in this map, the default connection will be used. Connections are not directional, so an error will be thrown if the same connection is defined twice (e.g. `Map[A][B] = someConnectionInfo`, and `Map[B][A] = otherConnectionInfo`).
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



PartitionConnectionInfo
-----------------------
This class is a plain old object defining the state between two partitions (e.g. whether network traffic is blocked or not). It is auto-generated from a gRPC API, so exploring it in code is the best way to view its properties. 

**NOTE:** These objects will often have several gRPC-specific fields inside them, but which don't need to be considered; you can construct the object however you normally instantiate objects in your language of choice (e.g. `new` in Java, `PartitionConnectionInfo{....fields...}` in Go, etc.).



ContainerConfig
---------------
Object containing information Kurtosis needs to create and run the container. This config should be created using [ContainerConfigBuilder][containerconfigbuilder] instances.

### String image
The name of the container image that Kurtosis should use when creating the service's container (e.g. `my-repo/my-image:some-tag-name`).

### Set\<String\> usedPortsSet
The set of ports that the container will be listening on, in the format `NUM/PROTOCOL` (e.g. `80/tcp`, `9090/udp`, etc.).

### Map\<String, String\> filesArtifactMountpoints
Sometimes a service needs files to be available before it starts, but creating those files manually is slow, difficult, or would require committing a very large artifact to the testsuite's Git repo (e.g. starting a service with a 5 GB Postgres database mounted). To ease this pain, Kurtosis allows you to specify URLs of gzipped TAR files that Kurtosis will download, uncompress, and mount inside your service containers. 

This property is therefore a map of the file artifact ID -> path on the container where the uncompressed artifact contents should be mounted, with the file artifact IDs corresponding to the files artifacts registered via [EnclaveContext.registerFilesArtifacts][enclavecontext_registerfilesartifacts]. 

E.g. if my test declares an artifact called `5gb-database` that lives at `https://my-site.com/test-artifacts/5gb-database.tgz`, I might return the following map from this function to mount the artifact at the `/database` path inside my container: `{"5gb-database": "/database"}`.

### List\<String\> entrypointOverrideArgs
You often won't control the container images that you'll be using in your testnet, and the `ENTRYPOINT` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### List\<String\> cmdOverrideArgs
You often won't control the container images that you'll be using in your testnet, and the `CMD` statement  hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements when necessary.

### Map\<String, String\> environmentVariableOverrides
Defines environment variables that should be set inside the Docker container running the service. This can be necessary for starting containers from Docker images you don't control, as they'll often be parameterized with environment variables.



ContainerConfigBuilder
------------------------------
The builder that should be used to create [ContainerConfig][containerconfig] instances. The functions on this builder will correspond to the properties on the [ContainerConfig][containerconfig] object, in the form `withPropertyName` (e.g. `withUsedPorts` sets the ports used by the container).



ServiceContext
--------------
This Kurtosis-provided class is the lowest-level representation of a service running inside a Docker container. It is your handle for retrieving container information and manipulating the container.

### getServiceId() -\> ServiceID
Gets the ID that Kurtosis uses to identify the service.

**Returns**

The service's ID.

### getIpAddress() -\> String
Gets the IP address of the Docker container that the service is running inside.

**Returns**

The service's IP address.

### getSharedDirectory() -\> [SharedPath][sharedpath]
Get the directory that is mounted on both the current container running this code and the service container, so that files can be passed back and forth. The directory is expressed as a [SharedPath][sharedpath] object, so file inside can be referenced by absolute filepath on either this container or the service contianer.

**Returns**

The [SharedPath][sharedpath] object.


### execCommand(List\<String\> command) -\> (int exitCode, String logs)
Uses [Docker exec](https://docs.docker.com/engine/reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**

* `command`: The args of the command to execute in the container.

**Returns**

* `exitCode`: The exit code of the command.
* `logs`: The output of the run command, assuming a UTF-8 encoding. **NOTE:** Commands that output non-UTF-8 output will likely be garbled!

SharedPath
----------
Simple structure that holds information about a filepath shared between two containers: this container, and a container running a service in a testnet. The actual object referenced by this path could be anything - a file, a directory, a symlink, nonexistent, etc.

### getAbsPathOnThisContainer() -\> String
For the object in the shared directory represented by this `SharedPath` object, gets the absolute filepath on the container where this code is running.

### getAbsPathOnServiceContainer() -\> String
For the object in the shared directory represented by this `SharedPath` object, gets the absolute filepath on the remote service container where the shared directory is also mounted.

### getChildPath(String relativePath) -\> [SharedPath][sharedpath]
Gets a new [SharedPath][sharedpath] object that represents another path inside the shared directory, relative to the current path object. E.g. if the shared directory had a subdirectory called `my-dir` which has `some-file.txt`, `sharedDirRoot.getChildPath("my-dir")` would represent that subdirectory and `sharedDirRoot.getChildPath("my-dir/some-file.txt")` would get the file inside.

**Args**

* `relativePath`: The relative path to add at the end of the [SharedPath][sharedpath].

**Returns**

The new [SharedPath][sharedpath] object.

---

_Found a bug? File it on [the repo](https://github.com/kurtosis-tech/kurtosis-client/issues)!_


<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[containerconfig]: #containerconfig
[containerconfig_usedports]: #setstring-usedportsset
[containerconfig_filesartifactmountpoints]: #mapstring-string-filesartifactmountpoints

[containerconfigbuilder]: #containerconfigbuilder

[modulecontext]: #modulecontext

[enclavecontext]: #enclavecontext
[enclavecontext_registerfilesartifacts]: #registerfilesartifactsmapfilesartifactid-string-filesartifacturls
[enclavecontext_addservice]: #addserviceserviceid-serviceid--funcstring-ipaddr-sharedpath-shareddirectory---containerconfig-containerconfigsupplier---servicecontext-servicecontext-mapstring-portbinding-hostportbindings
[enclavecontext_addservicetopartition]: #addservicetopartitionserviceid-serviceid-partitionid-partitionid-funcstring-ipaddr-sharedpath-shareddirectory---containerconfig-containerconfigsupplier---servicecontext-servicecontext-mapstring-portbinding-hostportbindings
[enclavecontext_repartitionnetwork]: #repartitionnetworkmappartitionid-setserviceid-partitionservices-mappartitionid-mappartitionid-partitionconnectioninfo-partitionconnections-partitionconnectioninfo-defaultconnection

[partitionconnectioninfo]: #partitionconnectioninfo

[servicecontext]: #servicecontext

[sharedpath]: #sharedpath
[sharedpath_getabspathonthiscontainer]: #getabspathonthiscontainer---string
[sharedpath_getabspathonservicecontainer]: #getabspathonservicecontainer---string

[test]: ../kurtosis-testsuite-api-lib/lib-documentation#testn-extends-network
[test_configure]: ../kurtosis-testsuite-api-lib/lib-documentation#configuretestconfigurationbuilder-builder
[test_setup]: ../kurtosis-testsuite-api-lib/lib-documentation#setupenclavecontext-enclavecontext---n
[test_run]: ../kurtosis-testsuite-api-lib/lib-documentation#runn-network
[test_gettestconfiguration]: ../kurtosis-testsuite-api-lib/lib-documentation#gettestconfiguration---testconfiguration

[testconfiguration]: ../kurtosis-testsuite-api-lib/lib-documentation#testconfiguration

[testconfigurationbuilder]: ../kurtosis-testsuite-api-lib/lib-documentation#testconfigurationbuilder

[testsuite]: ../kurtosis-testsuite-api-lib/lib-documentation#testsuite
