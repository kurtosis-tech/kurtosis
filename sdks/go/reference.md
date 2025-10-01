# Reference
## Engine
<details><summary><code>client.Engine.GetEngineInfo() -> *sdk.EngineInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.GetEngineInfo(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.ListEnclaves() -> map[string]*sdk.EnclaveInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.ListEnclaves(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.CreateEnclave(request) -> *sdk.EnclaveInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.CreateEnclave{
        EnclaveName: "enclave_name",
        ApiContainerVersionTag: "api_container_version_tag",
    }
client.Engine.CreateEnclave(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**apiContainerVersionTag:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**apiContainerLogLevel:** `*string` — Enclave log level, defaults to INFO
    
</dd>
</dl>

<dl>
<dd>

**mode:** `*sdk.EnclaveMode` — Enclave mode, defaults to TEST
    
</dd>
</dl>

<dl>
<dd>

**shouldApicRunInDebugMode:** `*sdk.ApiContainerDebugMode` — Whether the APIC's container should run with the debug server to receive a remote debug connection
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.DeleteEnclaves() -> *sdk.DeletionSummary</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete stopped enclaves. TO delete all the enclaves use the query parameter `remove_all` 
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.DeleteEnclavesRequest{
        RemoveAll: sdk.Bool(
            true,
        ),
    }
client.Engine.DeleteEnclaves(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**removeAll:** `*bool` — If true, remove all enclaves. Otherwise only remove stopped enclaves. Default is false
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.ListAllEnclaveIdentifiers() -> []*sdk.EnclaveIdentifiers</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.ListAllEnclaveIdentifiers(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.GetEnclaveDetailedInfo(EnclaveIdentifier) -> *sdk.EnclaveInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.GetEnclaveDetailedInfo(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.DestroyEnclave(EnclaveIdentifier) -> error</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.DestroyEnclave(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.GetEnclaveStatus(EnclaveIdentifier) -> *sdk.EnclaveStatus</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.GetEnclaveStatus(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Engine.SetEnclaveStatus(EnclaveIdentifier, request) -> error</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Engine.SetEnclaveStatus(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**request:** `sdk.EnclaveTargetStatus` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Enclave
<details><summary><code>client.Enclave.GetLastStarlarkRun(EnclaveIdentifier) -> *sdk.StarlarkDescription</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.GetLastStarlarkRun(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.UploadsAStarlarkPackage(EnclaveIdentifier, request) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.UploadsAStarlarkPackage(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ExecutesAStarlarkPackageOnTheUsersBehalf(EnclaveIdentifier, PackageId, request) -> *sdk.StarlarkRunResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

The endpoint will trigger the execution and deployment of a Starlark package. By default, it'll
return an async logs resource using `starlark_execution_uuid` that can be used to retrieve the logs
via streaming. It's also possible to block the call and wait for the execution to complete using the
query parameter `retrieve_logs_async`.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.RunStarlarkPackage{
        RetrieveLogsAsync: sdk.Bool(
            true,
        ),
    }
client.Enclave.ExecutesAStarlarkPackageOnTheUsersBehalf(
        context.TODO(),
        "enclave_identifier",
        "package_id",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**packageId:** `string` — The package identifier that will be executed
    
</dd>
</dl>

<dl>
<dd>

**retrieveLogsAsync:** `*bool` — If false, block http response until all logs are available. Default is true
    
</dd>
</dl>

<dl>
<dd>

**params:** `map[string]any` — Parameters data for the Starlark package main function
    
</dd>
</dl>

<dl>
<dd>

**dryRun:** `*bool` — Defaults to false
    
</dd>
</dl>

<dl>
<dd>

**parallelism:** `*int` — Defaults to 4
    
</dd>
</dl>

<dl>
<dd>

**clonePackage:** `*bool` 

Whether the package should be cloned or not.
If false, then the package will be pulled from the APIC local package store. If it's a local package then is must
have been uploaded using UploadStarlarkPackage prior to calling RunStarlarkPackage.
If true, then the package will be cloned from GitHub before execution starts
    
</dd>
</dl>

<dl>
<dd>

**relativePathToMainFile:** `*string` — The relative main file filepath, the default value is the "main.star" file in the root of a package
    
</dd>
</dl>

<dl>
<dd>

**mainFunctionName:** `*string` — The name of the main function, the default value is "run"
    
</dd>
</dl>

<dl>
<dd>

**experimentalFeatures:** `[]sdk.KurtosisFeatureFlag` 
    
</dd>
</dl>

<dl>
<dd>

**cloudInstanceId:** `*string` — Defaults to empty
    
</dd>
</dl>

<dl>
<dd>

**cloudUserId:** `*string` — Defaults to empty
    
</dd>
</dl>

<dl>
<dd>

**imageDownloadMode:** `*sdk.ImageDownloadMode` 
    
</dd>
</dl>

<dl>
<dd>

**nonBlockingMode:** `*bool` — Defaults to false
    
</dd>
</dl>

<dl>
<dd>

**githubAuthToken:** `*string` — Defaults to empty
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ExecutesAStarlarkScriptOnTheUsersBehalf(EnclaveIdentifier, request) -> *sdk.StarlarkRunResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

The endpoint will trigger the execution and deployment of a Starlark file. By default, it'll
return an async logs resource using `starlark_execution_uuid` that can be used to retrieve the logs
via streaming. It's also possible to block the call and wait for the execution to complete using the
query parameter `retrieve_logs_async`.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.RunStarlarkScript{
        RetrieveLogsAsync: sdk.Bool(
            true,
        ),
        SerializedScript: "serialized_script",
    }
client.Enclave.ExecutesAStarlarkScriptOnTheUsersBehalf(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**retrieveLogsAsync:** `*bool` — If false, block http response until all logs are available. Default is true
    
</dd>
</dl>

<dl>
<dd>

**serializedScript:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**params:** `map[string]any` — Parameters data for the Starlark package main function
    
</dd>
</dl>

<dl>
<dd>

**dryRun:** `*bool` — Defaults to false
    
</dd>
</dl>

<dl>
<dd>

**parallelism:** `*int` — Defaults to 4
    
</dd>
</dl>

<dl>
<dd>

**mainFunctionName:** `*string` — The name of the main function, the default value is "run"
    
</dd>
</dl>

<dl>
<dd>

**experimentalFeatures:** `[]sdk.KurtosisFeatureFlag` 
    
</dd>
</dl>

<dl>
<dd>

**cloudInstanceId:** `*string` — Defaults to empty
    
</dd>
</dl>

<dl>
<dd>

**cloudUserId:** `*string` — Defaults to empty
    
</dd>
</dl>

<dl>
<dd>

**imageDownloadMode:** `*sdk.ImageDownloadMode` 
    
</dd>
</dl>

<dl>
<dd>

**nonBlockingMode:** `*bool` — Defaults to false
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ReturnsDetailedInformationAboutASpecificService(EnclaveIdentifier, ServiceIdentifier) -> *sdk.ServiceInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.ReturnsDetailedInformationAboutASpecificService(
        context.TODO(),
        "enclave_identifier",
        "service_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceIdentifier:** `string` — The service identifier of the container that the command should be executed in
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ReturnsInformationAboutAllExistingHistoricalServices(EnclaveIdentifier) -> []*sdk.ServiceIdentifiers</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.ReturnsInformationAboutAllExistingHistoricalServices(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ReturnsDetailedInformationAboutAllsServicesWithinTheEnclave(EnclaveIdentifier) -> map[string]*sdk.ServiceInfo</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.GetEnclavesEnclaveIdentifierServicesRequest{}
client.Enclave.ReturnsDetailedInformationAboutAllsServicesWithinTheEnclave(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**services:** `*string` — Select services to get information
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ExecutesTheGivenCommandInsideARunningServicesContainer(EnclaveIdentifier, ServiceIdentifier, request) -> *sdk.ExecCommandResult</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.ExecCommand{
        CommandArgs: []string{
            "command_args",
        },
    }
client.Enclave.ExecutesTheGivenCommandInsideARunningServicesContainer(
        context.TODO(),
        "enclave_identifier",
        "service_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceIdentifier:** `string` — The service identifier of the container that the command should be executed in
    
</dd>
</dl>

<dl>
<dd>

**commandArgs:** `[]string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.CheckForServiceAvailability(EnclaveIdentifier, ServiceIdentifier, PortNumber) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Block until the given HTTP endpoint returns available, calling it through a HTTP request
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityRequest{
        HttpMethod: sdk.HttpMethodAvailabilityGet.Ptr(),
        Path: sdk.String(
            "path",
        ),
        InitialDelayMilliseconds: sdk.Int(
            1,
        ),
        Retries: sdk.Int(
            1,
        ),
        RetriesDelayMilliseconds: sdk.Int(
            1,
        ),
        ExpectedResponse: sdk.String(
            "expected_response",
        ),
        RequestBody: sdk.String(
            "request_body",
        ),
    }
client.Enclave.CheckForServiceAvailability(
        context.TODO(),
        "enclave_identifier",
        "service_identifier",
        1,
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceIdentifier:** `string` — The service identifier of the container that the command should be executed in
    
</dd>
</dl>

<dl>
<dd>

**portNumber:** `int` — The port number to check availability
    
</dd>
</dl>

<dl>
<dd>

**httpMethod:** `*sdk.HttpMethodAvailability` — The HTTP method used to check availability. Default is GET.
    
</dd>
</dl>

<dl>
<dd>

**path:** `*string` — The path of the service to check. It mustn't start with the first slash. For instance `service/health`
    
</dd>
</dl>

<dl>
<dd>

**initialDelayMilliseconds:** `*int` — The number of milliseconds to wait until executing the first HTTP call
    
</dd>
</dl>

<dl>
<dd>

**retries:** `*int` — Max number of HTTP call attempts that this will execute until giving up and returning an error
    
</dd>
</dl>

<dl>
<dd>

**retriesDelayMilliseconds:** `*int` — Number of milliseconds to wait between retries
    
</dd>
</dl>

<dl>
<dd>

**expectedResponse:** `*string` — If the endpoint returns this value, the service will be marked as available (e.g. Hello World).
    
</dd>
</dl>

<dl>
<dd>

**requestBody:** `*string` — If the http_method is set to POST, this value will be send as the body of the availability request.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.ListAllFilesArtifacts(EnclaveIdentifier) -> []*sdk.FileArtifactReference</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.ListAllFilesArtifacts(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.InspectTheContentOfAFileArtifact(EnclaveIdentifier, ArtifactIdentifier) -> []*sdk.FileArtifactDescription</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.InspectTheContentOfAFileArtifact(
        context.TODO(),
        "enclave_identifier",
        "artifact_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**artifactIdentifier:** `string` — The artifact name or uuid
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.UploadsLocalFileArtifactToTheKurtosisFileSystem(EnclaveIdentifier, request) -> map[string]*sdk.FileArtifactUploadResult</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.UploadsLocalFileArtifactToTheKurtosisFileSystem(
        context.TODO(),
        "enclave_identifier",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.AddRemoteFileToKurtosisFileSystem(EnclaveIdentifier, request) -> *sdk.FileArtifactReference</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Tells the API container to download a files artifact from the web to the Kurtosis File System
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.StoreWebFilesArtifact{
        Url: "url",
        Name: "name",
    }
client.Enclave.AddRemoteFileToKurtosisFileSystem(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**url:** `string` — URL to download the artifact from
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` — The name of the files artifact
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.AddServicesFileToKurtosisFileSystem(EnclaveIdentifier, ServiceIdentifier, request) -> *sdk.FileArtifactReference</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Tells the API container to copy a files artifact from a service to the Kurtosis File System
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.StoreFilesArtifactFromService{
        SourcePath: "source_path",
        Name: "name",
    }
client.Enclave.AddServicesFileToKurtosisFileSystem(
        context.TODO(),
        "enclave_identifier",
        "service_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceIdentifier:** `string` — The service identifier of the container that the command should be executed in
    
</dd>
</dl>

<dl>
<dd>

**sourcePath:** `string` — The absolute source path where the source files will be copied from
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` — The name of the files artifact
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Enclave.UserServicesPortForwarding(EnclaveIdentifier, request) -> error</code></summary>
<dl>
<dd>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Enclave.UserServicesPortForwarding(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**request:** `*sdk.Connect` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Streaming
<details><summary><code>client.Streaming.GetEnclavesServicesLogs(EnclaveIdentifier) -> *sdk.ServiceLogs</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get multiple enclave services logs concurrently. This endpoint can stream the logs by either starting
a Websocket connection (recommended) or legacy HTTP streaming.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.GetEnclavesEnclaveIdentifierLogsRequest{}
client.Streaming.GetEnclavesServicesLogs(
        context.TODO(),
        "enclave_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceUuidSet:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**followLogs:** `*bool` 
    
</dd>
</dl>

<dl>
<dd>

**conjunctiveFilters:** `*sdk.LogLineFilter` 
    
</dd>
</dl>

<dl>
<dd>

**returnAllLogs:** `*bool` 
    
</dd>
</dl>

<dl>
<dd>

**numLogLines:** `*int` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Streaming.GetServiceLogs(EnclaveIdentifier, ServiceIdentifier) -> *sdk.ServiceLogs</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get service logs. This endpoint can stream the logs by either starting
a Websocket connection (recommended) or legacy HTTP streaming.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &sdk.GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogsRequest{
        FollowLogs: sdk.Bool(
            true,
        ),
        ReturnAllLogs: sdk.Bool(
            true,
        ),
        NumLogLines: sdk.Int(
            1,
        ),
    }
client.Streaming.GetServiceLogs(
        context.TODO(),
        "enclave_identifier",
        "service_identifier",
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**enclaveIdentifier:** `string` — UUID, shortened UUID, or name of the enclave
    
</dd>
</dl>

<dl>
<dd>

**serviceIdentifier:** `string` — The service identifier of the container that the command should be executed in
    
</dd>
</dl>

<dl>
<dd>

**followLogs:** `*bool` 
    
</dd>
</dl>

<dl>
<dd>

**conjunctiveFilters:** `*sdk.LogLineFilter` 
    
</dd>
</dl>

<dl>
<dd>

**returnAllLogs:** `*bool` 
    
</dd>
</dl>

<dl>
<dd>

**numLogLines:** `*int` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Streaming.GetStarlarkExecutionLogs(StarlarkExecutionUuid) -> []*sdk.StarlarkRunResponseLine</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Stream the logs of an Starlark execution that were initiated using `retrieve_logs_async`.
The async logs can be consumed only once and expire after consumption or 2 hours after creation.
This endpoint can stream the logs by either starting a Websocket connection (recommended) or
legacy HTTP streaming.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Streaming.GetStarlarkExecutionLogs(
        context.TODO(),
        "starlark_execution_uuid",
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**starlarkExecutionUuid:** `string` — The unique identifier to track the execution of a Starlark script or package
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>
