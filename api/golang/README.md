# Kurtosis Go SDK

This guide provides instructions and code snippets to help you get started with the Kurtosis Go SDK. The SDK enables you to create and manage enclaves programmatically, without having to rely on Kurtosis UI or CLI.

## Setting Up

To use the Kurtosis Go SDK, you need to add it as a dependency to your Go module. You can do this with the following `go get` command:

```console
$ go get github.com/kurtosis-tech/kurtosis/api/golang
```

Please ensure that you have a running Kurtosis Engine instance before executing your code. You can check the status of the Kurtosis Engine using the following command:

```console
$ kurtosis engine status
```

Make note of the Engine's version and status information.

## Creating an Enclave

The first step is to obtain a Kurtosis Context, which represents a Kurtosis instance in your Go code:

```go
kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
```

Next, you can use the Kurtosis Context to create an enclave, which will provide an Enclave Context for managing the enclave:

```go
enclaveName := "my-enclave"
enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName)
```

Using the Enclave Context, you can perform actions like adding services using Starlark scripts:

```go
starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig()
starlarkScript := `
def run(plan):
    serviceConfig := ServiceConfig{
        Image: "httpd",
    }
    plan.AddService(name: "my-service", config: serviceConfig)
`
starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlarkRunConfig)
```

After adding a service, you can interact with it by obtaining a service context and running commands:

```go
serviceCtx, err := enclaveCtx.GetServiceContext("my-service")
code, output, err := serviceCtx.ExecCommand([]string{"ls"})
```

For ephemeral enclaves, such as those used in end-to-end testing, you can destroy the created enclave:

```go
err := kurtosisCtx.DestroyEnclave(ctx, enclaveName)
```

These instructions should help you get started with using the Kurtosis Go SDK to create and manage enclaves for your projects. If you need further assistance or ahve questions, please open a [Github Discussion](https://github.com/kurtosis-tech/kurtosis/discussions/categories/q-a) or ping us in [Discord](https://discord.com/invite/HUapYX9RvV).