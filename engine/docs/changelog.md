# TBD
### Features
* Try to pull `api-container` latest image before running the API container Docker container
* Upgraded to Kurt Core 1.25.1, which add `com.kurtosistech.app-id` container label to all enclave containers

# 0.1.0
### Features
* Added `EngineServerService` which will be in charge of receive requests for creating and destroying Kurtosis Enclaves
* Ported the `EnclaveManager` from `Kurtosis CLI` to here and it also includes `DockerNetworkAllocator`
* Added `MinimalRPCServer` used to implement the Kurtosis Engine RPC Server
* Added `build` script to automatize building Engine Server process
* Added `Docker` file for Kurtosis Engine Server Docker image
* Added `get-docker-image-tag` script to automatize Docker image tag creation process
* Added `.dockerignore` to enable Docker caching
* Added `Enclave` struct in `enclave_manager` packaged

### Changes
* Removed check Typescript job and publish Typescript artifact job in Circle CI configuration

### Removals
* Removed the `log` inside `EnclaveManager`, as it's no longer needed
