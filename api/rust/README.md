# Kurtosis Rust SDK

This is an SDK for [Kurtosis](https://github.com/kurtosis-tech/kurtosis), based on the protobufs available [here](https://github.com/kurtosis-tech/kurtosis/tree/main/api/protobuf);

## Example

Make sure that the engine is running:

```terminal
kurtosis engine start
```

Then you can run a Starlark script using Kurtosis+Rust:

```rust
use kurtosis_sdk::{engine_api::{engine_service_client::{EngineServiceClient}, CreateEnclaveArgs}, enclave_api::{api_container_service_client::ApiContainerServiceClient, RunStarlarkScriptArgs}};
use kurtosis_sdk::enclave_api::starlark_run_response_line::RunResponseLine::InstructionResult;

const STARLARK_SCRIPT : &str = "
def main(plan):
    plan.print('Hello World!')
";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // CREATE ENCLAVE
    let mut engine = EngineServiceClient::connect("https://[::1]:9710").await?;
    let create_enclave_response = engine.create_enclave(CreateEnclaveArgs{
        enclave_name: "my-rust-test".to_string(),
        api_container_log_level: "info".to_string(),
        // Default
        api_container_version_tag: "".to_string(),
        is_partitioning_enabled: false,
    }).await?.into_inner();
    
    // CONNECT TO ENCLAVE
    let enclave_port = create_enclave_response.enclave_info.expect("Enclave info must be present").api_container_host_machine_info.expect("Enclave host machine info must be present").grpc_port_on_host_machine;
    let mut enclave = ApiContainerServiceClient::connect(format!("https://[::1]:{}", enclave_port)).await?;
    
    // RUN STARLARK SCRIPT
    let mut run_result = enclave.run_starlark_script(RunStarlarkScriptArgs{
        serialized_script: STARLARK_SCRIPT.to_string(),
        serialized_params: "{}".to_string(),
        dry_run: Option::Some(false),
        parallelism: Option::None,
        main_function_name: "main".to_string(),
    }).await?.into_inner();
    
    // GET OUTPUT LINES
    while let Some(next_message) = run_result.message().await? {
        next_message.run_response_line.map(|line| match line {
            InstructionResult(result) => {
                println!("{}", result.serialized_instruction_result)
            }
            _ => (),
        });
    }
    Ok(())
}

```

More details can be found on the [docs](https://docs.kurtosis.com/).

