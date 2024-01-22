# Kurtosis Typescript SDK

This guide provides instructions and code snippets to help you get started with the Kurtosis Typescript SDK. It enables you to create and manage enclaves programmatically, without having to rely on the Kurtosis Enclave Manager (UI) or the Kurtosis CLI.

The main way to interact with objects from the Kurtosis ecosystem is by getting its *context*. There are three main contexts:
1. **KurtosisContext**: contains methods for interacting with Kurtosis Engine, allowing manipulation of enclaves.
2. **EnclaveContext**: contains methods for interacting with an enclave, allowing execution of Starlark scripts.
3. **ServiceContext**: contains methods for interacting with a service, allowing inspecting its details.

This guide will also help you create and get these contexts.

## Setting Up

To use the Kurtosis Typescript SDK, you need to add it as a dependency to your NPM package. You can do this with the following `npm i` command:

```console
$ npm i kurtosis-sdk
```

Please ensure that you have a running Kurtosis Engine instance before executing your code. You can check the status of the Kurtosis Engine using the following command:

```console
$ kurtosis engine status
```

Make note of the Engine's version and status information.

## Creating an Enclave

The first step is to obtain a Kurtosis Context, which represents a Kurtosis instance in your Typescript code:

```typescript
const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine()
if(newKurtosisContextResult.isErr()) {
    // Check for error
}
const kurtosisContext = newKurtosisContextResult.value;
```

Next, you can use the Kurtosis Context to create an enclave, which will provide an Enclave Context for managing the enclave:

```typescript
const enclaveName = "my-enclave"
const createEnclaveResult = await kurtosisContext.createEnclave();
if(createEnclaveResult.isErr()) {
    // Check for error
}
const enclaveContext = createEnclaveResult.value;
```

## Configure for Starlark Runs

Using the Enclave Context, you can perform actions like adding services using Starlark scripts:

```typescript
const starlarkRunConfig = new StarlarkRunConfig(
StarlarkRunConfig.WithSerializedParams(params)
)
const starlarkScript = `
def run(plan):
    serviceConfig := ServiceConfig{
        Image: "httpd",
    }
    plan.AddService(name: "my-service", config: serviceConfig)
`
return enclaveContext.runStarlarkScriptBlocking(starlarkScript, starlarkRunConfig)
```
## Interacting with services
After adding a service, you can interact with it by obtaining a service context and running commands:

```typescript
const getServiceCtxResult = await enclaveCtx.getServiceContext("my-service")
if(getServiceCtxResult.isErr()) {
    // Check for error
}
const serviceContext = getServiceCtxResult.value;
const command = ["ls"]
serviceContext.execCommand(command)
```

For ephemeral enclaves, such as those used in end-to-end testing, you can destroy the created enclave:

```typescript
const destroyEnclaveResponse = await kurtosisContext.destroyEnclave(enclaveName)
if(destroyEnclaveResponse.isErr()) {
    // Check for error
}
```

These instructions should help you get started with using the Kurtosis Typescript SDK to create and manage enclaves for your projects. If you need further assistance or have questions, please open a [Github Discussion](https://github.com/kurtosis-tech/kurtosis/discussions/categories/q-a) or ping us in [Discord](https://discord.com/invite/HUapYX9RvV).