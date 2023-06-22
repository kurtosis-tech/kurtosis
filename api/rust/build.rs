use std::io::Result;
fn main() -> Result<()> {
    tonic_build::compile_protos("../protobuf/engine/engine_service.proto")?;
    tonic_build::compile_protos("../protobuf/core/api_container_service.proto")?;
    Ok(())
}