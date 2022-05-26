# TBD
### Features
* Added a temporary parameter `publicPorts` in `KurtosisBackend.StartUserService` to support defining static public ports

### Breaking Changes
* Removed `FilesArtifactID` type because it is not used anywhere in this package
  * Users have to remove the dependency with this type and create their own type instead

# 0.32.0
### Changes
* Switched to using EmptyDir ephemeral volumes for the enclave data directory and the files artifact expansion volumes in response to learning that Kubernetes/DigitalOcean doesn't really want you creating lots of these

### Breaking Changes
* `GetEngineServerKubernetesKurtosisBackend` and `GetApiContainerKubernetesKurtosisBackend` no longer take in `storageClassName` or `enclaveSizeInMegabytes`
    * These parameters are no longer needed

# 0.31.1
### Fixes
* Fixed a bug in how we were checking for services which don't have pods yet

# 0.31.0
### Features
* Reworked how files artifact expansion works, so that our volumes only need to be mounted as `ReadWriteOnce`

### Breaking Changes
* Removed `FilesArtifactGUID` and the `FilesArtifactExpansion` types
* Removed the `KurtosisBackend.CreateFilesArtifactExpansion` and `DestroyFilesArtifactExpansions` methods
* `StartUserService` now takes in a map specifying a connection between files that get expanded on the files artifact expander container and mountpoints on the user service container


# 0.30.2
### Changes
* Use `kurtosis-enclave--UUID` for enclave namespace names
* `GetXXXXX` methods in Kubernetes manager only return non-tombstoned objects

### Fixes
* Fixed an issue when getting a single enclave object by its ID fails because a condition was wrong
* Fixed a bug related to using array of reference variables for `enclaveKubernetesResources.pods` and `enclaveKubernetesResources.services` fields

# 0.30.1
### Fixes
* Fixed an issue where a Service and a Pod could exist at the same time even though the Service won't have private port annotations

# 0.30.0
### Features
* Failed file artifact expansion jobs will now return their logs

### Changes
* Set pod restart policies to `Never` so that we have less magic going on

### Fixes
* Fixed issue with the files artifact expansion not waiting for the volume to get bound
* Gave the API container permission to create jobs
* Fixed issue with engine and API containers not having permissions to get Pod logs
* Fixed bug with the files artifact expansion job not mounting the enclave data volume
* Changed pod wait-for-availability timeout (how long the pod can stay in "Pending") from 1 minute to 15 minutes, because Pending also includes the time spent pulling images and some images can be very large (e.g. NEAR)
* Return an error if a pod's container hits ImagePullBackOff
* Fixed a bug where user services were mounting the same volume multiple times due to a reference to a for-loop variable (which will always have the value of the last iteration of the loop)
* Fixed a bug in declaring user service ports, due to the same reference-to-a-for-loop-variable thing

### Breaking Changes
* `CreateFilesArtifactExpansion` no longer takes in a `FilesArtifactID` (as it's unneeded)
    * Remediation: remove the argument in the call

# 0.29.1
### Changes
* Trying to run networking partitioning methods in Kubernetes will result in an error, rather than a panic
* Tidying up several things inside the codebase:
    * Kubernetes network partitioning methods now return an `error` rather than panicking
    * `PullImage` returns an error, rather than panicking, for both Docker & Kubernetes Kurtosis backends
    * Removed some dead code
* Remove synchronous deletes because they're too slow

# 0.29.0
### Features
* Implement Kubernetes-backed file artifact expansion
* Build `CopyFilesFromUserService` in Kubernetes
* Added a `main.go` that's Gitignored with some local testing structure already set up

### Changes
* Calls to remove Kubernetes resources are now synchronous

### Fixes
* Fix DockerLogStreamingReadCloser logging at ERROR level when it should log at DEBUG
* Ensured we're not going to get race conditions when writing the output of Docker & Kubernetes exec commands
* Fixed a bug in `getSingleUserServiceObjectAndKubernetesResources`

### Breaking Changes
* Renamed `KurtosisBackend.CopyFromUserService` -> `CopyFilesFromUserService`
    * Users should update their usages accordingly
* `KurtosisBackend.CopyFilesFromUserService` now copies all bytes synchronously, rather than returning a `ReadCloser` for the user to deal with
    * Remediation: pass in a `io.Writer` where the bytes should go

# 0.28.0
### Fixes
* Fixed bug related to having two annotations-key-consts for Kubernetes objects

### Breaking Changes
* `Module.GetPublicIP` is renamed to `GetMaybePublicIP`
    * Remediation: switch to new version
* `Module.GetPublicPorts` renamed to `GetMaybePublicPorts`
    * Remediation: switch to new version
* `Module.GetPublicIp` renamed to `Module.GetPublicIP`
    * Remediation: switch to new version
    
# 0.27.0
### Breaking Changes
* Unified file expansion volume and expanders into one interface with two associated methods (instead of two interfaces and four methods)

### Changes
* Switched the API container to get its port info from the serialized port specs on the Kubernetes service

### Fixes
* Fixed the engine container being listed as running if the engine service had selectors defined
* Switched our UUID generation to be fully random (v4) UUIDs rather than v1
* Fixed our exec command
* Fixed API containers and engines not being able to run pod exec commands
* Fixed a bug that caused service registration to fail
* Fixed the user services and modules Docker log streams not actually coming back demultiplexed

# 0.26.1
### Features
* Added `KubernetesKurtosisBackend.GetModuleLogs`

# 0.26.0
### Features
* Added the functionality to wait until the GRPC port is available before returning when creating `Engines`, `API containers` and `Modules`  
* Added `KubernetesKurtosisBackend.CreateModule`, `KubernetesKurtosisBackend.GetModules`, `KubernetesKurtosisBackend.StopModules` and `KubernetesKurtosisBackend.DestroyModules`
* Added `ForModulePod` and `ForModuleService` to `KubernetesEnclaveObjectAttributesProvider`
* Started proto-documentation on README about how the CRUD methods work, and why
* Switched user service objects to use UUIDs for service GUIDs
* Implement remaining user service methods:
    * `GetUserServices`
    * `StopUserServices`
    * `DestroyUserServices`
    
### Breaking Changes
* Removed `ModuleGUID` argument in `KurtosisBackend.CreateModule`
  * Users will need to remove the argument on each call, the module's GUID will be automatically created in the backend for them

### Changes
* Upgraded Kubernetes client SDK from v0.20 to v0.24
* Upgraded this library to depend on Go 1.17 (required for latest Kubernetes SDK)
* Switched the `UpdateService` implementation to use server-side apply

### Fixes
* Fix a bug in gathering user service Services and Pods
* Fix a nil pointer exception bug when starting a user service
* Fixes a bug with setting a user service's Service ports to empty if the user doesn't declare any ports

# 0.25.0
### Features
* Built out Kubernetes `GetUserServiceLogs`
* Built out Kubernetes `RunUserServiceExecCommands`

### Fixes
* Fixed `grpcProxy` port ID not being acceptable to Kubernetes
* Fixed a bug where RegisterService was creating Kubernetes Services without ports, which Kubernetes doesn't allow

### Changes
* The API container objects no longer get prefixed with the enclave name, and all get called `kurtosis-api` (which is fine because they're namespaced)

### Breaking Changes
* NewKubernetesManager now additionally takes in a Kubernetes configuration object
    * Users will need to pass in this new configuration (the same as is created when instantiating the Kubernetes clientset)
* Engine IDs are now of the formal type `EngineGUID` rather than `string`
    * All instances of engine ID strings need to be replaced with `EngineID` (e.g. all of the engine CRUD methods coming back from KurtosisBackend)
* The engine object's `GetID` method has now been renamed `GetGUID`
    * Users should switch to using the new method


### Breaking Changes
* Added the `enclaveId` argument in `GetModules`, `GetModuleLogs`, `StopModules` and `DestroyModules`
  * Users should add this new argument on each call

# 0.24.0
### Fixes
* Fixed a bug where the API container resources map would have entries even if the enclave was empty of API containers
* Fixed a bug where the API container didn't have a way to get enclave namespace names, because it isn't allowed to list namespaces given that its service account is a namespaced object

### Breaking Changes
* Renamed `GetLocalKubernetesKurtosisBackend` -> `GetCLIKubernetesKurtosisBackend`
    * Users should switch to the new version
* Split `GetInClusterKubernetesKurtosisBackend` -> `GetAPIContainerKubernetesKurtosisBackend` and `GetEngineServerKubernetesKurtosisBackend`
    * Users should select the right version appropriate to the user case

# 0.23.4
### Features
* Build out the following user service functions in Kubernetes:
    * `RegisterService`
    * `StartService`
    * All the pipeline for transforming Kubernetes objects into user services

### Fixes
* Fix bug in `waitForPersistentVolumeClaimBound` in which PVC name and namespace were flipped in args

### Changes
* Renamed all Kubernetes constants that were `XXXXXLabelKey` to be `XXXXXKubernetesLabelKey` to make it more visually obvious that we're using the Kubernetes constants rather than Docker
* Renamed all constants that were `XXXXXLabelValue` to be `XXXXXKubernetesLabelValue` to make it more visually obvious that we're using the Kubernetes constants rather than Docker
* Renamed all Docker constants that were `XXXXXLabelKey` to be `XXXXXDockerLabelKey` to make it more visually obvious that we're using the Docker constants rather than Kubernetes
* Renamed all constants that were `XXXXXLabelValue` to be `XXXXXDockerLabelValue` to make it more visually obvious that we're using the Docker constants rather than Kubernetes
* Renamed the Docker & Kubernetes port spec serializers to include their respective names, to be easier to visually identify in code

### Breaking Changes
* `StartUserService` now takes in a `map[FilesArtifactVolumeName]string` rather than `map[string]string` to be more explicit about the data it's consuming

### Fixes
* `KubernetesManager.CreatePod` now waits for pods to become available before returning

# 0.23.3
### Features
* Added `KubernetesKurtosisBackend.StopEnclaves` and `KubernetesKurtosisBackend.DestroyEnclaves`
* Added `KubernetesManager.IsPersistentVolumeClaimBound` to check when a Persistent Volume has been bound to a Persistent Volume Claim
* Updated `KubernetesManager.CreatePersistentVolumeClaim` to wait for the PersistentVolumeClaim to get bound
* Added `CollectMatchingRoles` and `CollectMatchingRoleBindings` in `kubernetes_resource_collectors` package
* Upped the CircleCI resource class to 'large' since builds are 1m30s and CircleCI showed that we're maxing out the CPU
* Added a build cache to each build
* Build out `KubernetesKurtosisBackend.DumpEnclave`

### Fixes
* Added apiv1 prefix to `KubernetesManager.GetPodPortforwardEndpointUrl`

### Fixes
* Added apiv1 prefix to `KubernetesManager.GetPodPortforwardEndpointUrl`

# 0.23.2
### Fixes
* Don't error when parsing public ports on stopped containers

# 0.23.1
### Fixes
* Fix accidentally calling pause when we should be unpausing
* Fix error-checking not getting checked after creating a service
* Fix bug with improper `nil` check in creating networking sidecar

# 0.23.0
### Breaking Changes
* For KubernetesBackends, `EnclaveVolumeSizeInGigabytes` has been changed from an int to a uint `EnclaveVolumeSizeInMegabytes`
  * Remediation: change all calls to KubernetesBackend factory methods to use megabytes (not gigabytes)

# 0.22.0
### Changes
* Upgrade to Docker SDK v20.10 to try and fix a bug where Docker network containers wouldn't be populated

### Fixes
* Fix an issue with network container IPs not being correctly made available when starting a DockerKurtosisBackend in API container mode
* Allowed removal of user service registrations independent of the underlying service registration (which is necessary for deregistering services in the API container)

### Removals
* Removed unneeded & unused `KurtosisBackend.WaitForEndpointAvailability` function

### Breaking Changes
* `user_service_registration.ServiceID` is now once again `service.ServiceID`
    * Users should update their code
* `user_service_registration.UserServiceRegistration` objects have been removed and replaced with `service.ServiceRegistration`
    * Users should use the new objects
* `user_service_registration.UserServiceRegistrationGUID` no longer exists
    * Users should switch to using `ServiceGUID`
* `CreateUserServiceRegistration` has been replaced with `RegisterUserService`
    * Users should use `RegisterUserService`
* `DestroyUserServiceRegistration` has been removed
    * Users should use `DestroyUserServices`
* `CreateUserService` has been renamed to `StartUserService`
    * Users should call `RegisterUserService` first, then `StartUserService`
* All user service functions now take in an `enclaveId` parameter
    * Users should provide the new parameter
* `CreateFilesArtifactExpansionVolume` now takes in a `ServiceGUID` once more
    * Users should switch back to providing a `ServiceGUID`

### Features
* Added `ForUserServiceService` for `KubernetesEngineObjectAttributesProvider`

# 0.21.1
### Fixes
* Add ID & GUID labels to enclave networks & namespaces

# 0.21.0
### Changes
* The `DockerKurtosisBackend` will now track the free IPs of networks
* `KurtosisBackend` now has `UserServiceRegistration` CRUD methods
* Service containers in Docker no longer get tagged with a service ID (this is now on the service registration object)

### Breaking Changes
* Renamed `service.ServiceID` to `user_service_registration.UserServiceID`
    * Users should update their imports/packages accordingly
* Removed the `ServiceID` field of the `Service` object
    * Users should query for the `UserServiceRegistration` associated with the service to find service ID
* `KurtosisBackend.CreateUserService` now takes in a service registration GUID, rather than a static IP
    * Users should make sure to call `CreateUserServiceRegistration` and pass in the returned GUID to the user service
* `GetLocalDockerKurtosisBackend` now takes in an optional argument for providing the enclave ID, if running inside an enclave
    * API containers should pass in the information; all other consumers of the `DockerKurtosisBackend` should pass in `nil`
* Removed the GUID parameter from `KurtosisBackend.CreateService`
    * Users no longer need to provide this parameter; it's autogenerated
* Switched the `CreateFilesArtifactExpansionVolume` from using `ServiceGUID` -> `UserServiceRegistrationGUID`
    * Users should now provide the registration GUID
* Removed the IP address argument from the following methods in `KurtosisBackend`:
    * `CreateAPIContainer`
    * `CreateModule`
    * `CreateService`
    * `CreateNetworkingSidecar`
    * `RunFilesArtifactExpander`
* `CreateAPIContainer` now takes in an extra `ownIpAddressEnvVar` environment variable, which is the environment variable the `KurtosisBackend` should populate with the API container's IP address
* Removed the following from the `Enclave` object:
    * `GetNetworkCIDR`
    * `GetNetworkID`
    * `GetNetworkGatewayIp`
    * `GetNetworkIpAddrTracker`

# 0.20.2

### Changes
* Updated Kurtosis Kubernetes `label_key_consts` and tests.
* Added `GetPodPortforwardEndpointUrl` method to kubernetes_manager

# 0.20.1
### Features
* Added `KubernetesKurtosisBackend.GetEnclaves` functionality to kubernetes backend
* Added `KubernetesKurtosisBackend.CreateAPIContainer`, `KubernetesKurtosisBackend.GetAPIContainers`, `KubernetesKurtosisBackend.StopAPIContainers` and `KubernetesKurtosisBackend.DestroyAPIContainers` 
* Added `KubernetesKurtosisBackend.isPodRunningDeterminer` utility variable that we use for determine if a pod is running
* Added `GetInClusterKubernetesKurtosisBackend` Kurtosis backend factory method to be used for pods inside Kubernetes cluster
* `Network` objects returned by `DockerManager` will have the gateway IP and the IPs of the containers on the network

# 0.20.0
### Features
* Added persistent volume claim creation to kubernetes-backed enclaves
* Added `CreateEnclave` functionality to kubernetes backend
* Added `ServiceAccounts`, `Roles`, `RoleBindings`, `ClusterRole`, and `ClusterRoleBindings` create, getByLabels and remove methods to `KubernetesManager`
* Added `ForEngineNamespace`, `ForEngineServiceAccount`, `ForEngineClusterRole` and `ForEngineClusterRoleBindings` to  `KubernetesEngineObjectAttributesProvider`
* Updated `KubernetesBackend.CreateEngine` added the kubernetes role based resources creation and namespace creation process
* Fixed `KubernetesBackend.GetEngines`returning an empty list for filters with no IDs specified
* Added a (currently unused) framework for collecting all Kubernetes resource that match a specific filter
* Add `getEngineKubernetesResources` in preparation for refactoring the engine methods
* Implement `KubernetesKurtosisBackend.DestroyEngines`

### Changes
* Updated `KubernetesManager.CreatePod` added `serviceAccount` argument to set the pod's service account
* Switched all the engine methods to use a more Kubernetes-friendly way of getting & managing resources
* Cleaned up the `KubernetesManager.CreateEngine` method

### Breaking Changes
* NewKurtosisKubernetesBackend now takes in extra arguments - `volumeStorageClassName` and `volumeSizePerEnclaveInGigabytes`

# 0.19.0
### Breaking Changes
* Removed `enclaveDataDirpathOnHostMachine` and `enclaveDataDirpathOnServiceContainer` from `KurtosisBackend.CreateUserService`
    * Users no longer need to provide this argument
* Removed `enclaveDataDirpathOnHostMachine` argument from `KurtosisBackend.CreateModule`
    * Users no longer need to provide this argument
* Removed `engineDataDirpathOnHostMachine` from `KurtosisBackend.CreateEngine`
    * Users no longer need to provide this argument

# 0.18.0
### Features
* Added `ServiceAccounts`, `Roles`, `RoleBindings`, `ClusterRole`, and `ClusterRoleBindings` create and remove methods to `KubernetesManager`
* Added `CreateEnclave` functionality to Kubernetes backend

### Changes
* Stopped mounting an enclave data directory on the API container

### Fixes
* `RunFilesArtifactExpander` now correctly only requires the user to pass in the filepath of the artifact to expand, relative to the enclave data volume root

### Breaking Changes
* Removed the `enclaveDataDirpathOnHostMachine` parameter from `KurtosisBackend.CreateAPIContainer`
    * Users no longer need to provide this parameter

# 0.17.0
### Features
* Added `PauseService` and `UnpauseService` to `KurtosisBackend`
* Added docker implementation of `PauseService` and `UnpauseService`
* Added Kubernetes implementation of engine functions in kubernetes backend

### Breaking Changes
* Added an extra `enclaveDataVolumeDirpath` to `KurtosisBackend.CreateAPIContainer`
    * Users should pass in the location where the enclave data volume should be mounted

# 0.16.0
### Removals
* Removed `files_artifact.FilesArtifactID` because it was a duplicate of `serivce.FilesArtifactID`

### Breaking Change
* Removed `files_artifact.FilesArtifactID`
    * Users should switch to `service.FilesArtifactID`

# 0.15.3
### Features
* Added `KurtosisBackend.CopyFromUserService` in Docker implementation

### Fixes
* Fixed a bug where module containers were getting duplicate mountpoints for enclave data volume & bindmounted dirpath

# 0.15.2
### Fixes
* Fix `DockerKurtosisBackend.getEnclaveDataVolumeByEnclaveId` helper method that was accidentally broken

# 0.15.1
### Features
* The enclave data volume gets mounted on all services
* Updated `DockerKurtosisBackend.CreateEnclave`, now also creates an enclave data volume
* Parallelized several operations to improve perf:
    * `DockerKurtosisBackend.StopEnclaves`
    * `DockerKurtosisBackend.DestroyEnclaves`
    * `DockerKurtosisBackend.StopAPIContainers`
    * `DockerKurtosisBackend.DestroyAPIContainers`
    * `DockerKurtosisBackend.StopEngines`
    * `DockerKurtosisBackend.DestroyEngines`
    * `DockerKurtosisBackend.StopModules`
    * `DockerKurtosisBackend.DestroyModules`
    * `DockerKurtosisBackend.StopNetworkingSidecars`
    * `DockerKurtosisBackend.DestroyNetworkingSidecars`
    * `DockerKurtosisBackend.StopUserServices`
    * `DockerKurtosisBackend.DestroyUserServices`

# 0.15.0
### Fixes
* Fixed `UserService` object not having a `GetPrivatePorts` method
* Fixed `UserService` to correctly have `nil` public ports if it's not running to match the spec

### Breaking Changes
* Renamed `UserService.GetPublicPorts` -> `UserService.GetMaybePublicPorts`
    * Users should rename the method call appropriately

# 0.14.5
### Features
* Removed enclave's volumes when executing `KurtosisBackend.DestroyEnclaves` in Docker implementation

# 0.14.4
### Fixes
* Fix the exec commands on user services & networking sidecars being wrapped in `sh`, leading to unintended consequences

# 0.14.3
### Fixes
* Temporarily support the old port spec (`portId.number-protocol_portId.number-protocol`) so that we're still backwards-compatible

# 0.14.2
### Fixes
* Added a check to prevent generate user service's public IP address and public ports if it does not contain host port bindings

# 0.14.1
### Fixes
* Fixed nil pointer dereference error when calling enclave's methods with nil filters argument
* Fixed instances of propagating nil errors
* Fixed ApiContainer GetPublicGRPCProxyPort returning PublicGRPCPort

# 0.14.0
### Features
* Added `FilesArtifactExpander` general object
* Added `KurtosisBackend.RunFilesArtifactExpander` and `KurtosisBackend.DestroyFilesArtifactExpanders`
* Added `DockerEnclaveObjectAttributesProvider.ForFilesArtifactExpanderContainer`
* Added `FilesArtifactExpansionVolume` general object
* Added `KurtosisBackend.CreateFilesArtifactExpansionVolume` and `KurtosisBackend.DestroyFilesArtifactExpansionVolumes`
* Added `DockerEnclaveObjectAttributesProvider.ForFilesArtifactExpansionVolume`

### Breaking Changes
* Change `DockerManager.GetVolumesByLabels` returned type, now it returns a `volume` Docker object list instead of a volume name list
  * Users can use the new returned object and get the name from it.

### Fixes
* Fixed nil pointer dereference error when sending nil `filter.IDs` value in `enclave` CRUD methods

# 0.13.0
### Breaking Changes
* Reverted the port specification delimeters back to new style
  * From: rpc.8545-TCP_ws.8546-TCP_tcpDiscovery.30303-TCP_udpDiscovery.30303-UDP
  * To:   rpc:8545/TCP,ws:8546/TCP,tcpDiscovery:30303/TCP,udpDiscovery:30303/UDP
  * Users should upgrade the Kurtosis Client.
  * Users should restart enclaves.

# 0.12.0
### Breaking Changes
* Reverted the port specification delimeters back to original style.
  * From: rpc:8545/TCP,ws:8546/TCP,tcpDiscovery:30303/TCP,udpDiscovery:30303/UDP
  * To:   rpc.8545-TCP_ws.8546-TCP_tcpDiscovery.30303-TCP_udpDiscovery.30303-UDP
  * Users should upgrade the Kurtosis Client.
  * Users should restart enclaves.
  
# 0.11.3
### Features
* Added `KurtosisBackend.GetModuleLogs`

### Fixes
* Fixed assigning entry to nil map in `killContainers` in `docker_kurtosis_backend`
* Set CGO_ENABLED=0 for tests. Complete tests without prompting for GCC-5 now.

# 0.11.1
### Features
* Added `KurtosisBackend.StopModules`

### Fixes
* Fixed a bug where we weren't being efficient when killing containers in the `StopXXXXXX` calls

# 0.11.0
### Features
* Added `ExecResult` object to represent the result of commands execution inside an instance

### Breaking Changes
* Replaced the first returned var type `map[service.ServiceGUID]bool` in `RunUserServiceExecCommands` and `RunNetworkingSidecarExecCommands` with `map[service.ServiceGUID]*exec_result.ExecResult`
  * Users can get the new returned map and use the `ExecResult` object to obtain the execution exit code and output 

# 0.10.4
### Features
* Implemented the methods: `CreateNetworkingSidecar`, `GetNetworkingSidecars`, `RunNetworkingSidecarExecCommands`, `StopNetworkingSidecars` and `DestroyNetworkingSidecars` in `DockerKurtosisBackend`
* Implemented the methods: `CreateUserService`, `GetUserServices`, `StopUserServices`, `GetUserServiceLogs` and `DestroyUserServices` in `DockerKurtosisBackend`
* Implemented the methods: `WaitForUserServiceHttpEndpointAvailability`, `GetConnectionWithUserService` and `RunUserServiceExecCommands` methods in `DockerKurtosisBackend`
* Added `NetworkingSidecarContainerTypeLabelValue` label value constant
* Added `UserServiceContainerTypeLabelValue` and `NetworkingSidecarContainerTypeLabelValue` label key constants

# 0.10.3
### Features
* Added module CRUD methods

### Fixes
* Fixed a bug in `CreateAPIContainer` where the wrong label was being used to check for enclave ID

# 0.10.2
### Features
* Added `UserService` methods, `Modules` methods and `CreateRepartition` method in `KurtosisBackend` interface
* Stubbing out methods for `UserService`, `Modules` and `CreateRepartition` into Docker implementation
* Expose information of an `APIContainer` object
* Implemented the methods: `CreateEnclave`, `GetEnclaves`, `StopEnclaves` and `DestroyEnclaves` in `DockerKurtosisBackend`
* Added `IsNetworkPartitioningEnabledLabelKey` label key constant
* Added `NetworkPartitioningEnabledLabelValue` and `NetworkPartitioningDisabledLabelValue` label value constants
* Added API container CRUD methods
* Added `KurtosisBackend.DumpEnclave`

# 0.10.1
### Features
* Added API container CRUD stubs to `KurtosisBackend`

# 0.10.0
### Features
* Created a generic `ContainerStatus` object in the `KurtosisBackend` API for representing the state of containers
* Added enclave CRUD commands to `KurtosisBackend`

### Breaking Changes
* The `Engine` object's status will now be a `ContainerStatus`, rather than `EngineStatus`
  * Users should migrate to the new version

# 0.9.1
### Fixes
* Fixed container `removing` state erroneously counting as running

# 0.9.0
### Changes
* Switched from fairly-specific engine methods on `KurtosisBackend` to generic CRUD methods (`CreateEngine`, `GetEngines`, etc.)
* Pull the `ForEngineServer` method into this repo, rather than relying on obj-attrs-schema-lib, since the name/label determination really makes sense next to the DockerKurtosisBackend
* Commented out the `KubernetesKurtosisBackend` until we can revisit it in the future
* The `DockerManager` now logs to the system `logrus` instance like everything else

### Removals
* Removed the `KurtosisXXXDrivers` classes, as we're going to use `KurtosisBackend` as the external-facing API instead

### Breaking Changes
* Removed the `KurtosisXXXDrivers` classes
  * Users should use the `KurtosisBackend` struct instead
* Changed the API of `KurtosisBackend` to have generic CRUD methods for engines
  * Users should adapt their previous uses of specific methods (like `Clean`) to instead use these generic CRUD APIs
* Switched the `ContainerStatus` enum to be an `int` with dmarkham/enumer autogenerated code
  * Users should switch to using the new methods autogenerated by the enumer code
* Removed the `logger` parameter to the `NewDockerManager` function, as it's no longer needed
  * Users should remove the parameter
* The `HostMachineDomainInsideContainer` and `HostGatewayName` parameters are no longer public
  * There is no replacement; users who were depending on this should use the `needsAccessToHostMachine` parameter when starting a container
* The `DockerManager` package has been moved
  * Users should replace `github.com/kurtosis-tech/container-engine-lib/lib/docker_manager` -> `github.com/kurtosis-tech/container-engine-lib/lib/backends/docker/docker_manager`


# 0.8.8
### Features
* Added the `KurtosisDockerDriver` struct with methods that don't do anything

# 0.8.7
### Features
* Added KubernetesManager, KurtosisBackendCore Interface with implementations for both kubernetes and docker and the KurtosisBackend layer
* Upgraded `object-attributes-schema-lib` to 0.7.2

# 0.8.6
### Fixes
* Run `go mod tidy`, to get rid of unnecessary `palantir/stacktrace` dependency

# 0.8.5
### Fixes
* `stacktrace.Propagate` now panics when it receives a `nil` input

# 0.8.4
### Features
* Container logs are propagated throughout the returning error when container start fails

# 0.8.3
### Fixes
* Fixed issue where `GetContainersByLabel` would return `0.0.0.0` addresses rather than `127.0.0.1`
* Replaced redundant `getContainerIdsByName` with `getContainersByFilterArgs`

# 0.8.2
### Fixes
* Host port bindings will be returned as `127.0.0.1` rather than `0.0.0.0`, because Windows machines don't automatically correct `0.0.0.0` to `127.0.0.1`

# 0.8.1
### Features
* Added `GetVolumesByLabels` and `GetNetworksByLabels` functions

### Fixes
* For comments on the public functions, added the function name as the first line of the comment to make GoLand stop complaining

# 0.8.0
### Features
* Added the ability to add labels to networks & volumes

### Breaking Changes
* `CreateNetwork` now also takes in a list of labels to give the network
* `CreateVolume` now also takes in a list of labels to give the volume

# 0.7.0
### Features
* Added the ability to specify fixed host machine port bindings when starting a container is now available via the key of the map in `CreateAndStartContainerArgsBuilder.WithUsedPorts` function

### Breaking Changes
* The `CreateAndStartContainerArgsBuilder.WithUsedPorts`'s parameter now has a value of `PortPublishSpec`, which defines how the port should be published
* `CreateAndStartContainerArgsBuilder.ShouldPublishAllPorts` parameter has been removed
  * Users should migrate to `CreateAndStartContainerArgsBuilder.WithUsedPorts` instead

# 0.6.1
### Features
* Added `RemoveVolume` and `RemoveContainer` functions
* Added a `GetVolumesByName` function

### Changes
* Clarified that the `all` argument to `GetContainersByLabels` is for whether stopped containers should be shown

# 0.6.0
### Features
* Add `Network` type to store Docker Network information as ID, name, ip and mask

### Breaking Changes
* Replaced `GetNetworkIdsByName` method with `GetNetworksByName` because it returns a `Network` list which offers more information
  * Users should replace `GetNetworkIdsByName` calls with `GetNetworksByName` and get the network ID and other network information from the `Network` list returned by this new method

# 0.5.0
### Breaking Changes
* Renamed `Status` type to `ContainerStatus` which is used in the `Container` struct
  * Users should replace `Status` type with `ContainerStatus` in all places where it being used

### Features
* Add `HostPortBindings` field into `Container` type to store the container public ports

# 0.4.4
### Changes
* Removes Docker container name prefix `/` before using it to set a Container's name

# 0.4.3
### Features
* Add `Container` type to store container's information as ID, names, labels and status
* Add `GetContainersByLabels` method to get a container list by labels

### Changes
* Removes `GetContainerIdsByLabels` it'll be replaced by `GetContainersByLabels`which returns container ids as well along with other container info

### Fixes
* Fixed a bug where `CreateAndStartContainerArgsBuilder.Build` method wasn't setting container labels

# 0.4.2
### Fixes
* Fixed a bug where the defer function in `CreateAndStartContainer` would try to kill an empty container ID

# 0.4.1
### Features
* Add container's labels configuration when a container is started
* Add a new method `DockerManager.GetContainerIdsByLabels` to get a container ids list by container's labels

# 0.4.0
### Features
* `CreateAndStartContainerArgsBuilder`'s boolean-setting functions now accept a boolean

### Breaking Changes
* `CreateAndStartContainerArgsBuilder`'s boolean-setting functions now accept a boolean (used to be no-arg)

# 0.3.0
### Features
* `CreateAndStartContainer` now accepts args as built by a builder (hallelujah!)

### Breaking Changes
* `CreateAndStartContainer` now accepts a `CreateAndStartContainerArgs` object, rather than the gigantic list of parameters

# 0.2.10
### Fixes
* Fixed some logging that was being incorrectly being done through `logrus`, rather than `DockerManager.log`

# 0.2.9
### Fixes
* Added retry logic when trying to get host port bindings for a container, to account for https://github.com/moby/moby/issues/42860

# 0.2.8
### Features
* Made `PullImage` a public function on `DockerManager`

# 0.2.7
### Features
* Added extra trace logging to Docker manager

# 0.2.6
### Changes
* Added extra debug logging to hunt down an issue where the `defer` to remove a container that doesn't start properly would try to remove a container with an empty container ID (which would return a "not found" from the Docker engine)

# 0.2.5
### Fixes
* Fixed a bug where not specifying an image tag (which should default to `latest`) wouldn't actually pull the image if it didn't exist locally

# 0.2.4
### Fixes
* Add extra error-checking to handle a very weird case we just saw where container creation succeeds but no container ID is allocated

# 0.2.3
### Features
* Verify that, when starting a container with `shouldPublishAllPorts` == `true`, each used port gets exactly one host machine port binding

### Fixes
* Fixed ports not getting bound when running a container with an `EXPOSE` directive in the image Dockerfile
* Fixed broken CI

# 0.2.2
### Fixes
* Fixed bug in markdown-link-check property

# 0.2.1
### Features
* Set up CircleCI checks

# 0.2.0
### Breaking Changes
* Added an `alias` field to `CreateAndStartContainer`

# 0.1.0
* Init commit
