/// ==============================================================================================
///                            Shared Objects (Used By Multiple Endpoints)
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Port {
    #[prost(uint32, tag = "1")]
    pub number: u32,
    /// The protocol that the port is listening on
    #[prost(enumeration = "port::TransportProtocol", tag = "2")]
    pub transport_protocol: i32,
    #[prost(string, tag = "3")]
    pub maybe_application_protocol: ::prost::alloc::string::String,
    /// The wait timeout duration in string
    #[prost(string, tag = "4")]
    pub maybe_wait_timeout: ::prost::alloc::string::String,
}
/// Nested message and enum types in `Port`.
pub mod port {
    #[derive(
        Clone,
        Copy,
        Debug,
        PartialEq,
        Eq,
        Hash,
        PartialOrd,
        Ord,
        ::prost::Enumeration
    )]
    #[repr(i32)]
    pub enum TransportProtocol {
        Tcp = 0,
        Sctp = 1,
        Udp = 2,
    }
    impl TransportProtocol {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                TransportProtocol::Tcp => "TCP",
                TransportProtocol::Sctp => "SCTP",
                TransportProtocol::Udp => "UDP",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "TCP" => Some(Self::Tcp),
                "SCTP" => Some(Self::Sctp),
                "UDP" => Some(Self::Udp),
                _ => None,
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Container {
    #[prost(enumeration = "container::Status", tag = "1")]
    pub status: i32,
    #[prost(string, tag = "2")]
    pub image_name: ::prost::alloc::string::String,
    #[prost(string, repeated, tag = "3")]
    pub entrypoint_args: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(string, repeated, tag = "4")]
    pub cmd_args: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(map = "string, string", tag = "5")]
    pub env_vars: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ::prost::alloc::string::String,
    >,
}
/// Nested message and enum types in `Container`.
pub mod container {
    #[derive(
        Clone,
        Copy,
        Debug,
        PartialEq,
        Eq,
        Hash,
        PartialOrd,
        Ord,
        ::prost::Enumeration
    )]
    #[repr(i32)]
    pub enum Status {
        Stopped = 0,
        Running = 1,
        Unknown = 2,
    }
    impl Status {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Status::Stopped => "STOPPED",
                Status::Running => "RUNNING",
                Status::Unknown => "UNKNOWN",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "STOPPED" => Some(Self::Stopped),
                "RUNNING" => Some(Self::Running),
                "UNKNOWN" => Some(Self::Unknown),
                _ => None,
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ServiceInfo {
    /// UUID of the service
    #[prost(string, tag = "1")]
    pub service_uuid: ::prost::alloc::string::String,
    /// The IP address of the service inside the enclave
    #[prost(string, tag = "2")]
    pub private_ip_addr: ::prost::alloc::string::String,
    /// The ports on which the service is reachable inside the enclave, specified in user_specified_port_id -> port_info
    /// Will be exactly what was passed in at the time of starting the service
    #[prost(map = "string, message", tag = "3")]
    pub private_ports: ::std::collections::HashMap<::prost::alloc::string::String, Port>,
    /// Public IP address *outside* the enclave where the service is reachable
    /// NOTE: Will be empty if the service isn't running, the service didn't define any ports, or the backend doesn't support reporting public service info
    #[prost(string, tag = "4")]
    pub maybe_public_ip_addr: ::prost::alloc::string::String,
    /// Mapping defining the ports that the service can be reached at *outside* the enclave, in the user_defined_port_id -> port_info where user_defined_port_id
    ///   corresponds to the ID that was passed in in AddServiceArgs
    /// NOTE: Will be empty if the service isn't running, the service didn't define any ports, or the backend doesn't support reporting public service info
    #[prost(map = "string, message", tag = "5")]
    pub maybe_public_ports: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        Port,
    >,
    /// Name of the service
    #[prost(string, tag = "6")]
    pub name: ::prost::alloc::string::String,
    /// Shortened uuid of the service
    #[prost(string, tag = "7")]
    pub shortened_uuid: ::prost::alloc::string::String,
    /// Service status: stopped, running.
    #[prost(enumeration = "ServiceStatus", tag = "8")]
    pub service_status: i32,
    /// Docker container or Kubernetes pod container
    #[prost(message, optional, tag = "9")]
    pub container: ::core::option::Option<Container>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RunStarlarkScriptArgs {
    #[prost(string, tag = "1")]
    pub serialized_script: ::prost::alloc::string::String,
    #[prost(string, optional, tag = "2")]
    pub serialized_params: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to false
    #[prost(bool, optional, tag = "3")]
    pub dry_run: ::core::option::Option<bool>,
    /// Defaults to 4
    #[prost(int32, optional, tag = "4")]
    pub parallelism: ::core::option::Option<i32>,
    /// The name of the main function, the default value is "run"
    #[prost(string, optional, tag = "5")]
    pub main_function_name: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(enumeration = "KurtosisFeatureFlag", repeated, tag = "6")]
    pub experimental_features: ::prost::alloc::vec::Vec<i32>,
    /// Defaults to empty
    #[prost(string, optional, tag = "7")]
    pub cloud_instance_id: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to empty
    #[prost(string, optional, tag = "8")]
    pub cloud_user_id: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to empty
    #[prost(enumeration = "ImageDownloadMode", optional, tag = "9")]
    pub image_download_mode: ::core::option::Option<i32>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RunStarlarkPackageArgs {
    #[prost(string, tag = "1")]
    pub package_id: ::prost::alloc::string::String,
    /// Serialized parameters data for the Starlark package main function
    /// This should be a valid JSON string
    #[prost(string, optional, tag = "5")]
    pub serialized_params: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to false
    #[prost(bool, optional, tag = "6")]
    pub dry_run: ::core::option::Option<bool>,
    /// Defaults to 4
    #[prost(int32, optional, tag = "7")]
    pub parallelism: ::core::option::Option<i32>,
    /// Whether the package should be cloned or not.
    /// If false, then the package will be pulled from the APIC local package store. If it's a local package then is must
    /// have been uploaded using UploadStarlarkPackage prior to calling RunStarlarkPackage.
    /// If true, then the package will be cloned from GitHub before execution starts
    #[prost(bool, optional, tag = "8")]
    pub clone_package: ::core::option::Option<bool>,
    /// The relative main file filepath, the default value is the "main.star" file in the root of a package
    #[prost(string, optional, tag = "9")]
    pub relative_path_to_main_file: ::core::option::Option<
        ::prost::alloc::string::String,
    >,
    /// The name of the main function, the default value is "run"
    #[prost(string, optional, tag = "10")]
    pub main_function_name: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(enumeration = "KurtosisFeatureFlag", repeated, tag = "11")]
    pub experimental_features: ::prost::alloc::vec::Vec<i32>,
    /// Defaults to empty
    /// Deprecated: This value isn't used in the APIC anymore
    #[prost(string, optional, tag = "12")]
    pub cloud_instance_id: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to empty
    /// Deprecated: This value isn't used in the APIC anymore
    #[prost(string, optional, tag = "13")]
    pub cloud_user_id: ::core::option::Option<::prost::alloc::string::String>,
    /// Defaults to empty
    #[prost(enumeration = "ImageDownloadMode", optional, tag = "14")]
    pub image_download_mode: ::core::option::Option<i32>,
    /// Deprecated: If the package is local, it should have been uploaded with UploadStarlarkPackage prior to calling
    /// RunStarlarkPackage. If the package is remote and must be cloned within the APIC, use the standalone boolean flag
    /// clone_package below
    #[prost(oneof = "run_starlark_package_args::StarlarkPackageContent", tags = "3, 4")]
    pub starlark_package_content: ::core::option::Option<
        run_starlark_package_args::StarlarkPackageContent,
    >,
}
/// Nested message and enum types in `RunStarlarkPackageArgs`.
pub mod run_starlark_package_args {
    /// Deprecated: If the package is local, it should have been uploaded with UploadStarlarkPackage prior to calling
    /// RunStarlarkPackage. If the package is remote and must be cloned within the APIC, use the standalone boolean flag
    /// clone_package below
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum StarlarkPackageContent {
        /// the payload of the local module
        #[prost(bytes, tag = "3")]
        Local(::prost::alloc::vec::Vec<u8>),
        /// just a flag to indicate the module must be cloned inside the API
        #[prost(bool, tag = "4")]
        Remote(bool),
    }
}
/// ==============================================================================================
///                                Starlark Execution Response
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkRunResponseLine {
    #[prost(
        oneof = "starlark_run_response_line::RunResponseLine",
        tags = "1, 2, 3, 4, 5, 6, 7"
    )]
    pub run_response_line: ::core::option::Option<
        starlark_run_response_line::RunResponseLine,
    >,
}
/// Nested message and enum types in `StarlarkRunResponseLine`.
pub mod starlark_run_response_line {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum RunResponseLine {
        #[prost(message, tag = "1")]
        Instruction(super::StarlarkInstruction),
        #[prost(message, tag = "2")]
        Error(super::StarlarkError),
        #[prost(message, tag = "3")]
        ProgressInfo(super::StarlarkRunProgress),
        #[prost(message, tag = "4")]
        InstructionResult(super::StarlarkInstructionResult),
        #[prost(message, tag = "5")]
        RunFinishedEvent(super::StarlarkRunFinishedEvent),
        #[prost(message, tag = "6")]
        Warning(super::StarlarkWarning),
        #[prost(message, tag = "7")]
        Info(super::StarlarkInfo),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInfo {
    #[prost(string, tag = "1")]
    pub info_message: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkWarning {
    #[prost(string, tag = "1")]
    pub warning_message: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInstruction {
    #[prost(message, optional, tag = "1")]
    pub position: ::core::option::Option<StarlarkInstructionPosition>,
    #[prost(string, tag = "2")]
    pub instruction_name: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "3")]
    pub arguments: ::prost::alloc::vec::Vec<StarlarkInstructionArg>,
    #[prost(string, tag = "4")]
    pub executable_instruction: ::prost::alloc::string::String,
    #[prost(bool, tag = "5")]
    pub is_skipped: bool,
    #[prost(string, tag = "6")]
    pub description: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInstructionResult {
    #[prost(string, tag = "1")]
    pub serialized_instruction_result: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInstructionArg {
    #[prost(string, tag = "1")]
    pub serialized_arg_value: ::prost::alloc::string::String,
    #[prost(string, optional, tag = "2")]
    pub arg_name: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(bool, tag = "3")]
    pub is_representative: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInstructionPosition {
    #[prost(string, tag = "1")]
    pub filename: ::prost::alloc::string::String,
    #[prost(int32, tag = "2")]
    pub line: i32,
    #[prost(int32, tag = "3")]
    pub column: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkError {
    #[prost(oneof = "starlark_error::Error", tags = "1, 2, 3")]
    pub error: ::core::option::Option<starlark_error::Error>,
}
/// Nested message and enum types in `StarlarkError`.
pub mod starlark_error {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Error {
        #[prost(message, tag = "1")]
        InterpretationError(super::StarlarkInterpretationError),
        #[prost(message, tag = "2")]
        ValidationError(super::StarlarkValidationError),
        #[prost(message, tag = "3")]
        ExecutionError(super::StarlarkExecutionError),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkInterpretationError {
    #[prost(string, tag = "1")]
    pub error_message: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkValidationError {
    #[prost(string, tag = "1")]
    pub error_message: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkExecutionError {
    #[prost(string, tag = "1")]
    pub error_message: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkRunProgress {
    #[prost(string, repeated, tag = "1")]
    pub current_step_info: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(uint32, tag = "2")]
    pub total_steps: u32,
    #[prost(uint32, tag = "3")]
    pub current_step_number: u32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StarlarkRunFinishedEvent {
    #[prost(bool, tag = "1")]
    pub is_run_successful: bool,
    #[prost(string, optional, tag = "2")]
    pub serialized_output: ::core::option::Option<::prost::alloc::string::String>,
}
/// ==============================================================================================
///                                           Get Services
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetServicesArgs {
    /// "Set" of identifiers to fetch info for
    /// If empty, will fetch info for all services
    #[prost(map = "string, bool", tag = "1")]
    pub service_identifiers: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        bool,
    >,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetServicesResponse {
    /// "Set" from identifiers -> info about the service
    #[prost(map = "string, message", tag = "1")]
    pub service_info: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ServiceInfo,
    >,
}
/// An service identifier is a collection of uuid, name and shortened uuid
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ServiceIdentifiers {
    /// UUID of the service
    #[prost(string, tag = "1")]
    pub service_uuid: ::prost::alloc::string::String,
    /// Name of the service
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    /// The shortened uuid of the service
    #[prost(string, tag = "3")]
    pub shortened_uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetExistingAndHistoricalServiceIdentifiersResponse {
    #[prost(message, repeated, tag = "1")]
    pub all_identifiers: ::prost::alloc::vec::Vec<ServiceIdentifiers>,
}
/// ==============================================================================================
///                                           Exec Command
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecCommandArgs {
    /// The service identifier of the container that the command should be executed in
    #[prost(string, tag = "1")]
    pub service_identifier: ::prost::alloc::string::String,
    #[prost(string, repeated, tag = "2")]
    pub command_args: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecCommandResponse {
    #[prost(int32, tag = "1")]
    pub exit_code: i32,
    /// Assumes UTF-8 encoding
    #[prost(string, tag = "2")]
    pub log_output: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                              Wait For HTTP Get Endpoint Availability
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WaitForHttpGetEndpointAvailabilityArgs {
    /// The identifier of the service to check.
    #[prost(string, tag = "1")]
    pub service_identifier: ::prost::alloc::string::String,
    /// The port of the service to check. For instance 8080
    #[prost(uint32, tag = "2")]
    pub port: u32,
    /// The path of the service to check. It mustn't start with the first slash. For instance `service/health`
    #[prost(string, optional, tag = "3")]
    pub path: ::core::option::Option<::prost::alloc::string::String>,
    /// The number of milliseconds to wait until executing the first HTTP call
    #[prost(uint32, optional, tag = "4")]
    pub initial_delay_milliseconds: ::core::option::Option<u32>,
    /// Max number of HTTP call attempts that this will execute until giving up and returning an error
    #[prost(uint32, optional, tag = "5")]
    pub retries: ::core::option::Option<u32>,
    /// Number of milliseconds to wait between retries
    #[prost(uint32, optional, tag = "6")]
    pub retries_delay_milliseconds: ::core::option::Option<u32>,
    /// If the endpoint returns this value, the service will be marked as available (e.g. Hello World).
    #[prost(string, optional, tag = "7")]
    pub body_text: ::core::option::Option<::prost::alloc::string::String>,
}
/// ==============================================================================================
///                            Wait For HTTP Post Endpoint Availability
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WaitForHttpPostEndpointAvailabilityArgs {
    /// The identifier of the service to check.
    #[prost(string, tag = "1")]
    pub service_identifier: ::prost::alloc::string::String,
    /// The port of the service to check. For instance 8080
    #[prost(uint32, tag = "2")]
    pub port: u32,
    /// The path of the service to check. It mustn't start with the first slash. For instance `service/health`
    #[prost(string, optional, tag = "3")]
    pub path: ::core::option::Option<::prost::alloc::string::String>,
    /// The content of the request body.
    #[prost(string, optional, tag = "4")]
    pub request_body: ::core::option::Option<::prost::alloc::string::String>,
    /// The number of milliseconds to wait until executing the first HTTP call
    #[prost(uint32, optional, tag = "5")]
    pub initial_delay_milliseconds: ::core::option::Option<u32>,
    /// Max number of HTTP call attempts that this will execute until giving up and returning an error
    #[prost(uint32, optional, tag = "6")]
    pub retries: ::core::option::Option<u32>,
    /// Number of milliseconds to wait between retries
    #[prost(uint32, optional, tag = "7")]
    pub retries_delay_milliseconds: ::core::option::Option<u32>,
    /// If the endpoint returns this value, the service will be marked as available (e.g. Hello World).
    #[prost(string, optional, tag = "8")]
    pub body_text: ::core::option::Option<::prost::alloc::string::String>,
}
/// ==============================================================================================
///                                           Streamed Data Chunk
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StreamedDataChunk {
    /// Chunk of the overall files artifact bytes
    #[prost(bytes = "vec", tag = "1")]
    pub data: ::prost::alloc::vec::Vec<u8>,
    /// Hash of the PREVIOUS chunk, or empty string is this is the first chunk
    /// Referencing the previous chunk via its hash allows Kurtosis to validate
    /// the consistency of the data in case some chunk were not received
    #[prost(string, tag = "2")]
    pub previous_chunk_hash: ::prost::alloc::string::String,
    /// Additional metadata about the item being streamed
    #[prost(message, optional, tag = "3")]
    pub metadata: ::core::option::Option<DataChunkMetadata>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DataChunkMetadata {
    #[prost(string, tag = "1")]
    pub name: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                           Upload Files Artifact
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UploadFilesArtifactResponse {
    /// UUID of the files artifact, for use when referencing it in the future
    #[prost(string, tag = "1")]
    pub uuid: ::prost::alloc::string::String,
    /// UUID of the files artifact, for use when referencing it in the future
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                           Download Files Artifact
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DownloadFilesArtifactArgs {
    /// Files identifier to get bytes for
    #[prost(string, tag = "1")]
    pub identifier: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                         Store Web Files Artifact
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StoreWebFilesArtifactArgs {
    /// URL to download the artifact from
    #[prost(string, tag = "1")]
    pub url: ::prost::alloc::string::String,
    /// The name of the files artifact
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StoreWebFilesArtifactResponse {
    /// UUID of the files artifact, for use when referencing it in the future
    #[prost(string, tag = "1")]
    pub uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StoreFilesArtifactFromServiceArgs {
    /// Identifier that will be used to identify the service where the source files will be copied from
    #[prost(string, tag = "1")]
    pub service_identifier: ::prost::alloc::string::String,
    /// The absolute source path where the source files will be copied from
    #[prost(string, tag = "2")]
    pub source_path: ::prost::alloc::string::String,
    /// The name of the files artifact
    #[prost(string, tag = "3")]
    pub name: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StoreFilesArtifactFromServiceResponse {
    /// UUID of the files artifact, for use when referencing it in the future
    #[prost(string, tag = "1")]
    pub uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FilesArtifactNameAndUuid {
    /// A string representing the name of the file
    #[prost(string, tag = "1")]
    pub file_name: ::prost::alloc::string::String,
    /// A string representing the uuid of the file
    #[prost(string, tag = "2")]
    pub file_uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListFilesArtifactNamesAndUuidsResponse {
    #[prost(message, repeated, tag = "1")]
    pub file_names_and_uuids: ::prost::alloc::vec::Vec<FilesArtifactNameAndUuid>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct InspectFilesArtifactContentsRequest {
    #[prost(message, optional, tag = "1")]
    pub file_names_and_uuid: ::core::option::Option<FilesArtifactNameAndUuid>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct InspectFilesArtifactContentsResponse {
    #[prost(message, repeated, tag = "1")]
    pub file_descriptions: ::prost::alloc::vec::Vec<FileArtifactContentsFileDescription>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FileArtifactContentsFileDescription {
    /// Path relative to the file artifact
    #[prost(string, tag = "1")]
    pub path: ::prost::alloc::string::String,
    /// Size of the file, in bytes
    #[prost(uint64, tag = "2")]
    pub size: u64,
    /// A bit of text content, if the file allows (similar to UNIX's 'head')
    #[prost(string, optional, tag = "3")]
    pub text_preview: ::core::option::Option<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectServicesArgs {
    #[prost(enumeration = "Connect", tag = "1")]
    pub connect: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectServicesResponse {}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetStarlarkRunResponse {
    #[prost(string, tag = "1")]
    pub package_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub serialized_script: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub serialized_params: ::prost::alloc::string::String,
    #[prost(int32, tag = "4")]
    pub parallelism: i32,
    #[prost(string, tag = "5")]
    pub relative_path_to_main_file: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub main_function_name: ::prost::alloc::string::String,
    #[prost(enumeration = "KurtosisFeatureFlag", repeated, tag = "7")]
    pub experimental_features: ::prost::alloc::vec::Vec<i32>,
    #[prost(enumeration = "RestartPolicy", tag = "8")]
    pub restart_policy: i32,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ServiceStatus {
    Stopped = 0,
    Running = 1,
    Unknown = 2,
}
impl ServiceStatus {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ServiceStatus::Stopped => "STOPPED",
            ServiceStatus::Running => "RUNNING",
            ServiceStatus::Unknown => "UNKNOWN",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "STOPPED" => Some(Self::Stopped),
            "RUNNING" => Some(Self::Running),
            "UNKNOWN" => Some(Self::Unknown),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ImageDownloadMode {
    Always = 0,
    Missing = 1,
}
impl ImageDownloadMode {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ImageDownloadMode::Always => "always",
            ImageDownloadMode::Missing => "missing",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "always" => Some(Self::Always),
            "missing" => Some(Self::Missing),
            _ => None,
        }
    }
}
/// User services port forwarding
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Connect {
    /// Best effort port forwarding
    Connect = 0,
    /// Port forwarding disabled
    ///
    /// Starlark run fails if the ports cannot be forwarded.
    /// MUST_CONNECT = 2;
    NoConnect = 1,
}
impl Connect {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Connect::Connect => "CONNECT",
            Connect::NoConnect => "NO_CONNECT",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "CONNECT" => Some(Self::Connect),
            "NO_CONNECT" => Some(Self::NoConnect),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum KurtosisFeatureFlag {
    NoInstructionsCaching = 0,
}
impl KurtosisFeatureFlag {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            KurtosisFeatureFlag::NoInstructionsCaching => "NO_INSTRUCTIONS_CACHING",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NO_INSTRUCTIONS_CACHING" => Some(Self::NoInstructionsCaching),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum RestartPolicy {
    Never = 0,
    Always = 1,
}
impl RestartPolicy {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            RestartPolicy::Never => "NEVER",
            RestartPolicy::Always => "ALWAYS",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NEVER" => Some(Self::Never),
            "ALWAYS" => Some(Self::Always),
            _ => None,
        }
    }
}
/// Generated client implementations.
pub mod api_container_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct ApiContainerServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ApiContainerServiceClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> ApiContainerServiceClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> ApiContainerServiceClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
            >>::Error: Into<StdError> + Send + Sync,
        {
            ApiContainerServiceClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        /// Executes a Starlark script on the user's behalf
        pub async fn run_starlark_script(
            &mut self,
            request: impl tonic::IntoRequest<super::RunStarlarkScriptArgs>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::StarlarkRunResponseLine>>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/RunStarlarkScript",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "RunStarlarkScript",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /// Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
        pub async fn upload_starlark_package(
            &mut self,
            request: impl tonic::IntoStreamingRequest<Message = super::StreamedDataChunk>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status> {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/UploadStarlarkPackage",
            );
            let mut req = request.into_streaming_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "UploadStarlarkPackage",
                    ),
                );
            self.inner.client_streaming(req, path, codec).await
        }
        /// Executes a Starlark script on the user's behalf
        pub async fn run_starlark_package(
            &mut self,
            request: impl tonic::IntoRequest<super::RunStarlarkPackageArgs>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::StarlarkRunResponseLine>>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/RunStarlarkPackage",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "RunStarlarkPackage",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /// Returns the IDs of the current services in the enclave
        pub async fn get_services(
            &mut self,
            request: impl tonic::IntoRequest<super::GetServicesArgs>,
        ) -> std::result::Result<
            tonic::Response<super::GetServicesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/GetServices",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "GetServices",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Returns information about all existing & historical services
        pub async fn get_existing_and_historical_service_identifiers(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetExistingAndHistoricalServiceIdentifiersResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "GetExistingAndHistoricalServiceIdentifiers",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Executes the given command inside a running container
        pub async fn exec_command(
            &mut self,
            request: impl tonic::IntoRequest<super::ExecCommandArgs>,
        ) -> std::result::Result<
            tonic::Response<super::ExecCommandResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/ExecCommand",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "ExecCommand",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
        pub async fn wait_for_http_get_endpoint_availability(
            &mut self,
            request: impl tonic::IntoRequest<
                super::WaitForHttpGetEndpointAvailabilityArgs,
            >,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status> {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "WaitForHttpGetEndpointAvailability",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
        pub async fn wait_for_http_post_endpoint_availability(
            &mut self,
            request: impl tonic::IntoRequest<
                super::WaitForHttpPostEndpointAvailabilityArgs,
            >,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status> {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "WaitForHttpPostEndpointAvailability",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Uploads a files artifact to the Kurtosis File System
        pub async fn upload_files_artifact(
            &mut self,
            request: impl tonic::IntoStreamingRequest<Message = super::StreamedDataChunk>,
        ) -> std::result::Result<
            tonic::Response<super::UploadFilesArtifactResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/UploadFilesArtifact",
            );
            let mut req = request.into_streaming_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "UploadFilesArtifact",
                    ),
                );
            self.inner.client_streaming(req, path, codec).await
        }
        /// Downloads a files artifact from the Kurtosis File System
        pub async fn download_files_artifact(
            &mut self,
            request: impl tonic::IntoRequest<super::DownloadFilesArtifactArgs>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::StreamedDataChunk>>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/DownloadFilesArtifact",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "DownloadFilesArtifact",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /// Tells the API container to download a files artifact from the web to the Kurtosis File System
        pub async fn store_web_files_artifact(
            &mut self,
            request: impl tonic::IntoRequest<super::StoreWebFilesArtifactArgs>,
        ) -> std::result::Result<
            tonic::Response<super::StoreWebFilesArtifactResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/StoreWebFilesArtifact",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "StoreWebFilesArtifact",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Tells the API container to copy a files artifact from a service to the Kurtosis File System
        pub async fn store_files_artifact_from_service(
            &mut self,
            request: impl tonic::IntoRequest<super::StoreFilesArtifactFromServiceArgs>,
        ) -> std::result::Result<
            tonic::Response<super::StoreFilesArtifactFromServiceResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/StoreFilesArtifactFromService",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "StoreFilesArtifactFromService",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_files_artifact_names_and_uuids(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::ListFilesArtifactNamesAndUuidsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "ListFilesArtifactNamesAndUuids",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn inspect_files_artifact_contents(
            &mut self,
            request: impl tonic::IntoRequest<super::InspectFilesArtifactContentsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::InspectFilesArtifactContentsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/InspectFilesArtifactContents",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "InspectFilesArtifactContents",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// User services port forwarding
        pub async fn connect_services(
            &mut self,
            request: impl tonic::IntoRequest<super::ConnectServicesArgs>,
        ) -> std::result::Result<
            tonic::Response<super::ConnectServicesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/ConnectServices",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "ConnectServices",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Get last Starlark run
        pub async fn get_starlark_run(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetStarlarkRunResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.ApiContainerService/GetStarlarkRun",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "api_container_api.ApiContainerService",
                        "GetStarlarkRun",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod api_container_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with ApiContainerServiceServer.
    #[async_trait]
    pub trait ApiContainerService: Send + Sync + 'static {
        /// Server streaming response type for the RunStarlarkScript method.
        type RunStarlarkScriptStream: futures_core::Stream<
                Item = std::result::Result<super::StarlarkRunResponseLine, tonic::Status>,
            >
            + Send
            + 'static;
        /// Executes a Starlark script on the user's behalf
        async fn run_starlark_script(
            &self,
            request: tonic::Request<super::RunStarlarkScriptArgs>,
        ) -> std::result::Result<
            tonic::Response<Self::RunStarlarkScriptStream>,
            tonic::Status,
        >;
        /// Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
        async fn upload_starlark_package(
            &self,
            request: tonic::Request<tonic::Streaming<super::StreamedDataChunk>>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status>;
        /// Server streaming response type for the RunStarlarkPackage method.
        type RunStarlarkPackageStream: futures_core::Stream<
                Item = std::result::Result<super::StarlarkRunResponseLine, tonic::Status>,
            >
            + Send
            + 'static;
        /// Executes a Starlark script on the user's behalf
        async fn run_starlark_package(
            &self,
            request: tonic::Request<super::RunStarlarkPackageArgs>,
        ) -> std::result::Result<
            tonic::Response<Self::RunStarlarkPackageStream>,
            tonic::Status,
        >;
        /// Returns the IDs of the current services in the enclave
        async fn get_services(
            &self,
            request: tonic::Request<super::GetServicesArgs>,
        ) -> std::result::Result<
            tonic::Response<super::GetServicesResponse>,
            tonic::Status,
        >;
        /// Returns information about all existing & historical services
        async fn get_existing_and_historical_service_identifiers(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetExistingAndHistoricalServiceIdentifiersResponse>,
            tonic::Status,
        >;
        /// Executes the given command inside a running container
        async fn exec_command(
            &self,
            request: tonic::Request<super::ExecCommandArgs>,
        ) -> std::result::Result<
            tonic::Response<super::ExecCommandResponse>,
            tonic::Status,
        >;
        /// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
        async fn wait_for_http_get_endpoint_availability(
            &self,
            request: tonic::Request<super::WaitForHttpGetEndpointAvailabilityArgs>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status>;
        /// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
        async fn wait_for_http_post_endpoint_availability(
            &self,
            request: tonic::Request<super::WaitForHttpPostEndpointAvailabilityArgs>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status>;
        /// Uploads a files artifact to the Kurtosis File System
        async fn upload_files_artifact(
            &self,
            request: tonic::Request<tonic::Streaming<super::StreamedDataChunk>>,
        ) -> std::result::Result<
            tonic::Response<super::UploadFilesArtifactResponse>,
            tonic::Status,
        >;
        /// Server streaming response type for the DownloadFilesArtifact method.
        type DownloadFilesArtifactStream: futures_core::Stream<
                Item = std::result::Result<super::StreamedDataChunk, tonic::Status>,
            >
            + Send
            + 'static;
        /// Downloads a files artifact from the Kurtosis File System
        async fn download_files_artifact(
            &self,
            request: tonic::Request<super::DownloadFilesArtifactArgs>,
        ) -> std::result::Result<
            tonic::Response<Self::DownloadFilesArtifactStream>,
            tonic::Status,
        >;
        /// Tells the API container to download a files artifact from the web to the Kurtosis File System
        async fn store_web_files_artifact(
            &self,
            request: tonic::Request<super::StoreWebFilesArtifactArgs>,
        ) -> std::result::Result<
            tonic::Response<super::StoreWebFilesArtifactResponse>,
            tonic::Status,
        >;
        /// Tells the API container to copy a files artifact from a service to the Kurtosis File System
        async fn store_files_artifact_from_service(
            &self,
            request: tonic::Request<super::StoreFilesArtifactFromServiceArgs>,
        ) -> std::result::Result<
            tonic::Response<super::StoreFilesArtifactFromServiceResponse>,
            tonic::Status,
        >;
        async fn list_files_artifact_names_and_uuids(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::ListFilesArtifactNamesAndUuidsResponse>,
            tonic::Status,
        >;
        async fn inspect_files_artifact_contents(
            &self,
            request: tonic::Request<super::InspectFilesArtifactContentsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::InspectFilesArtifactContentsResponse>,
            tonic::Status,
        >;
        /// User services port forwarding
        async fn connect_services(
            &self,
            request: tonic::Request<super::ConnectServicesArgs>,
        ) -> std::result::Result<
            tonic::Response<super::ConnectServicesResponse>,
            tonic::Status,
        >;
        /// Get last Starlark run
        async fn get_starlark_run(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetStarlarkRunResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct ApiContainerServiceServer<T: ApiContainerService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: ApiContainerService> ApiContainerServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            let inner = _Inner(inner);
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for ApiContainerServiceServer<T>
    where
        T: ApiContainerService,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            let inner = self.inner.clone();
            match req.uri().path() {
                "/api_container_api.ApiContainerService/RunStarlarkScript" => {
                    #[allow(non_camel_case_types)]
                    struct RunStarlarkScriptSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::ServerStreamingService<super::RunStarlarkScriptArgs>
                    for RunStarlarkScriptSvc<T> {
                        type Response = super::StarlarkRunResponseLine;
                        type ResponseStream = T::RunStarlarkScriptStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RunStarlarkScriptArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).run_starlark_script(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = RunStarlarkScriptSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/UploadStarlarkPackage" => {
                    #[allow(non_camel_case_types)]
                    struct UploadStarlarkPackageSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::ClientStreamingService<super::StreamedDataChunk>
                    for UploadStarlarkPackageSvc<T> {
                        type Response = ();
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                tonic::Streaming<super::StreamedDataChunk>,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).upload_starlark_package(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = UploadStarlarkPackageSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.client_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/RunStarlarkPackage" => {
                    #[allow(non_camel_case_types)]
                    struct RunStarlarkPackageSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::ServerStreamingService<
                        super::RunStarlarkPackageArgs,
                    > for RunStarlarkPackageSvc<T> {
                        type Response = super::StarlarkRunResponseLine;
                        type ResponseStream = T::RunStarlarkPackageStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RunStarlarkPackageArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).run_starlark_package(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = RunStarlarkPackageSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/GetServices" => {
                    #[allow(non_camel_case_types)]
                    struct GetServicesSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<super::GetServicesArgs>
                    for GetServicesSvc<T> {
                        type Response = super::GetServicesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetServicesArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).get_services(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetServicesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers" => {
                    #[allow(non_camel_case_types)]
                    struct GetExistingAndHistoricalServiceIdentifiersSvc<
                        T: ApiContainerService,
                    >(
                        pub Arc<T>,
                    );
                    impl<T: ApiContainerService> tonic::server::UnaryService<()>
                    for GetExistingAndHistoricalServiceIdentifiersSvc<T> {
                        type Response = super::GetExistingAndHistoricalServiceIdentifiersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner)
                                    .get_existing_and_historical_service_identifiers(request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetExistingAndHistoricalServiceIdentifiersSvc(
                            inner,
                        );
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/ExecCommand" => {
                    #[allow(non_camel_case_types)]
                    struct ExecCommandSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<super::ExecCommandArgs>
                    for ExecCommandSvc<T> {
                        type Response = super::ExecCommandResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ExecCommandArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).exec_command(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ExecCommandSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability" => {
                    #[allow(non_camel_case_types)]
                    struct WaitForHttpGetEndpointAvailabilitySvc<T: ApiContainerService>(
                        pub Arc<T>,
                    );
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<
                        super::WaitForHttpGetEndpointAvailabilityArgs,
                    > for WaitForHttpGetEndpointAvailabilitySvc<T> {
                        type Response = ();
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::WaitForHttpGetEndpointAvailabilityArgs,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner)
                                    .wait_for_http_get_endpoint_availability(request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = WaitForHttpGetEndpointAvailabilitySvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability" => {
                    #[allow(non_camel_case_types)]
                    struct WaitForHttpPostEndpointAvailabilitySvc<
                        T: ApiContainerService,
                    >(
                        pub Arc<T>,
                    );
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<
                        super::WaitForHttpPostEndpointAvailabilityArgs,
                    > for WaitForHttpPostEndpointAvailabilitySvc<T> {
                        type Response = ();
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::WaitForHttpPostEndpointAvailabilityArgs,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner)
                                    .wait_for_http_post_endpoint_availability(request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = WaitForHttpPostEndpointAvailabilitySvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/UploadFilesArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct UploadFilesArtifactSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::ClientStreamingService<super::StreamedDataChunk>
                    for UploadFilesArtifactSvc<T> {
                        type Response = super::UploadFilesArtifactResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                tonic::Streaming<super::StreamedDataChunk>,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).upload_files_artifact(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = UploadFilesArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.client_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/DownloadFilesArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct DownloadFilesArtifactSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::ServerStreamingService<
                        super::DownloadFilesArtifactArgs,
                    > for DownloadFilesArtifactSvc<T> {
                        type Response = super::StreamedDataChunk;
                        type ResponseStream = T::DownloadFilesArtifactStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DownloadFilesArtifactArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).download_files_artifact(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = DownloadFilesArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/StoreWebFilesArtifact" => {
                    #[allow(non_camel_case_types)]
                    struct StoreWebFilesArtifactSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<super::StoreWebFilesArtifactArgs>
                    for StoreWebFilesArtifactSvc<T> {
                        type Response = super::StoreWebFilesArtifactResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::StoreWebFilesArtifactArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).store_web_files_artifact(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = StoreWebFilesArtifactSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/StoreFilesArtifactFromService" => {
                    #[allow(non_camel_case_types)]
                    struct StoreFilesArtifactFromServiceSvc<T: ApiContainerService>(
                        pub Arc<T>,
                    );
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<
                        super::StoreFilesArtifactFromServiceArgs,
                    > for StoreFilesArtifactFromServiceSvc<T> {
                        type Response = super::StoreFilesArtifactFromServiceResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::StoreFilesArtifactFromServiceArgs,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).store_files_artifact_from_service(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = StoreFilesArtifactFromServiceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids" => {
                    #[allow(non_camel_case_types)]
                    struct ListFilesArtifactNamesAndUuidsSvc<T: ApiContainerService>(
                        pub Arc<T>,
                    );
                    impl<T: ApiContainerService> tonic::server::UnaryService<()>
                    for ListFilesArtifactNamesAndUuidsSvc<T> {
                        type Response = super::ListFilesArtifactNamesAndUuidsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).list_files_artifact_names_and_uuids(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListFilesArtifactNamesAndUuidsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/InspectFilesArtifactContents" => {
                    #[allow(non_camel_case_types)]
                    struct InspectFilesArtifactContentsSvc<T: ApiContainerService>(
                        pub Arc<T>,
                    );
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<
                        super::InspectFilesArtifactContentsRequest,
                    > for InspectFilesArtifactContentsSvc<T> {
                        type Response = super::InspectFilesArtifactContentsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::InspectFilesArtifactContentsRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).inspect_files_artifact_contents(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = InspectFilesArtifactContentsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/ConnectServices" => {
                    #[allow(non_camel_case_types)]
                    struct ConnectServicesSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<
                        T: ApiContainerService,
                    > tonic::server::UnaryService<super::ConnectServicesArgs>
                    for ConnectServicesSvc<T> {
                        type Response = super::ConnectServicesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ConnectServicesArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).connect_services(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ConnectServicesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.ApiContainerService/GetStarlarkRun" => {
                    #[allow(non_camel_case_types)]
                    struct GetStarlarkRunSvc<T: ApiContainerService>(pub Arc<T>);
                    impl<T: ApiContainerService> tonic::server::UnaryService<()>
                    for GetStarlarkRunSvc<T> {
                        type Response = super::GetStarlarkRunResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).get_starlark_run(request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetStarlarkRunSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        Ok(
                            http::Response::builder()
                                .status(200)
                                .header("grpc-status", "12")
                                .header("content-type", "application/grpc")
                                .body(empty_body())
                                .unwrap(),
                        )
                    })
                }
            }
        }
    }
    impl<T: ApiContainerService> Clone for ApiContainerServiceServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    impl<T: ApiContainerService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: ApiContainerService> tonic::server::NamedService
    for ApiContainerServiceServer<T> {
        const NAME: &'static str = "api_container_api.ApiContainerService";
    }
}
