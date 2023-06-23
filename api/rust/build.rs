use std::io::Result;
use std::str;
use std::process::Command;

fn main() -> Result<()> {
    // Protobuf compilation only happens inside Kurtosis monorepo
    let regenerate_protobuf_bindings = option_env!("KURTOSIS_REGENERATE_BINDINGS");
    if regenerate_protobuf_bindings.is_some() {
        let mut kurtosis_folder_command = Command::new("git");
        kurtosis_folder_command.arg("rev-parse").arg("--show-toplevel");
        let kurtosis_folder = kurtosis_folder_command.output()?;
        let kurtosis_folder_str = str::from_utf8(&kurtosis_folder.stdout).unwrap().trim();
        return tonic_build::configure()
            .out_dir("src")
            .compile(&[
                format!("{}/api/protobuf/engine/engine_service.proto", kurtosis_folder_str),
                format!("{}/api/protobuf/core/api_container_service.proto", kurtosis_folder_str)
            ], &[kurtosis_folder_str])
    }
    Ok(())
}