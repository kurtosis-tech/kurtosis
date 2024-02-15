/// ==============================================================================================
///                                         Get Engine Info
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetEngineInfoResponse {
    /// Version of the engine server
    #[prost(string, tag = "1")]
    pub engine_version: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                         Create Enclave
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateEnclaveArgs {
    /// The name of the new Kurtosis Enclave
    #[prost(string, optional, tag = "1")]
    pub enclave_name: ::core::option::Option<::prost::alloc::string::String>,
    /// The image tag of the API container that should be used inside the enclave
    /// If blank, will use the default version that the engine server uses
    #[prost(string, optional, tag = "2")]
    pub api_container_version_tag: ::core::option::Option<
        ::prost::alloc::string::String,
    >,
    /// The API container log level
    #[prost(string, optional, tag = "3")]
    pub api_container_log_level: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(enumeration = "EnclaveMode", optional, tag = "4")]
    pub mode: ::core::option::Option<i32>,
    /// Whether the APIC's container should run with the debug server to receive a remote debug connection
    /// This is not an EnclaveMode because we will need to debug both current Modes (Test and Prod)
    #[prost(bool, optional, tag = "5")]
    pub should_apic_run_in_debug_mode: ::core::option::Option<bool>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateEnclaveResponse {
    /// All the enclave information inside this object
    #[prost(message, optional, tag = "1")]
    pub enclave_info: ::core::option::Option<EnclaveInfo>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnclaveApiContainerInfo {
    /// The container engine ID of the API container
    #[prost(string, tag = "1")]
    pub container_id: ::prost::alloc::string::String,
    /// The IP inside the enclave network of the API container (i.e. how services inside the network can reach the API container)
    #[prost(string, tag = "2")]
    pub ip_inside_enclave: ::prost::alloc::string::String,
    /// The grpc port inside the enclave network that the API container is listening on
    #[prost(uint32, tag = "3")]
    pub grpc_port_inside_enclave: u32,
    /// this is the bridge ip address that gets assigned to api container
    #[prost(string, tag = "6")]
    pub bridge_ip_address: ::prost::alloc::string::String,
}
/// Will only be present if the API container is running
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnclaveApiContainerHostMachineInfo {
    /// The interface IP on the container engine host machine where the API container can be reached
    #[prost(string, tag = "4")]
    pub ip_on_host_machine: ::prost::alloc::string::String,
    /// The grpc port on the container engine host machine where the API container can be reached
    #[prost(uint32, tag = "5")]
    pub grpc_port_on_host_machine: u32,
}
/// Enclaves are defined by a network in the container system, which is why there's a bunch of network information here
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnclaveInfo {
    /// UUID of the enclave
    #[prost(string, tag = "1")]
    pub enclave_uuid: ::prost::alloc::string::String,
    /// Name of the enclave
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    /// The shortened uuid of the enclave
    #[prost(string, tag = "3")]
    pub shortened_uuid: ::prost::alloc::string::String,
    /// State of all containers in the enclave
    #[prost(enumeration = "EnclaveContainersStatus", tag = "4")]
    pub containers_status: i32,
    /// State specifically of the API container
    #[prost(enumeration = "EnclaveApiContainerStatus", tag = "5")]
    pub api_container_status: i32,
    /// NOTE: Will not be present if the API container status is "NONEXISTENT"!!
    #[prost(message, optional, tag = "6")]
    pub api_container_info: ::core::option::Option<EnclaveApiContainerInfo>,
    /// NOTE: Will not be present if the API container status is not "RUNNING"!!
    #[prost(message, optional, tag = "7")]
    pub api_container_host_machine_info: ::core::option::Option<
        EnclaveApiContainerHostMachineInfo,
    >,
    /// The enclave's creation time
    #[prost(message, optional, tag = "8")]
    pub creation_time: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(enumeration = "EnclaveMode", tag = "9")]
    pub mode: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetEnclavesResponse {
    /// Mapping of enclave_uuid -> info_about_enclave
    #[prost(map = "string, message", tag = "1")]
    pub enclave_info: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        EnclaveInfo,
    >,
}
/// An enclave identifier is a collection of uuid, name and shortened uuid
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnclaveIdentifiers {
    /// UUID of the enclave
    #[prost(string, tag = "1")]
    pub enclave_uuid: ::prost::alloc::string::String,
    /// Name of the enclave
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    /// The shortened uuid of the enclave
    #[prost(string, tag = "3")]
    pub shortened_uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetExistingAndHistoricalEnclaveIdentifiersResponse {
    #[prost(message, repeated, tag = "1")]
    pub all_identifiers: ::prost::alloc::vec::Vec<EnclaveIdentifiers>,
}
/// ==============================================================================================
///                                        Stop Enclave
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StopEnclaveArgs {
    /// The identifier(uuid, shortened uuid, name) of the Kurtosis enclave to stop
    #[prost(string, tag = "1")]
    pub enclave_identifier: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                        Destroy Enclave
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DestroyEnclaveArgs {
    /// The identifier(uuid, shortened uuid, name) of the Kurtosis enclave to destroy
    #[prost(string, tag = "1")]
    pub enclave_identifier: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                        Create Enclave
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CleanArgs {
    /// If true, It will clean even the running enclaves
    #[prost(bool, optional, tag = "1")]
    pub should_clean_all: ::core::option::Option<bool>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EnclaveNameAndUuid {
    #[prost(string, tag = "1")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub uuid: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CleanResponse {
    /// removed enclave name and uuids
    #[prost(message, repeated, tag = "1")]
    pub removed_enclave_name_and_uuids: ::prost::alloc::vec::Vec<EnclaveNameAndUuid>,
}
/// ==============================================================================================
///                                    Get User Service Logs
/// ==============================================================================================
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetServiceLogsArgs {
    /// The identifier of the user service's Kurtosis Enclave
    #[prost(string, tag = "1")]
    pub enclave_identifier: ::prost::alloc::string::String,
    /// "Set" of service UUIDs in the enclave
    #[prost(map = "string, bool", tag = "2")]
    pub service_uuid_set: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        bool,
    >,
    /// If true, It will follow the container logs
    #[prost(bool, optional, tag = "3")]
    pub follow_logs: ::core::option::Option<bool>,
    /// The conjunctive log lines filters, the first filter is applied over the found log lines, the second filter is applied over the filter one result and so on (like grep)
    #[prost(message, repeated, tag = "4")]
    pub conjunctive_filters: ::prost::alloc::vec::Vec<LogLineFilter>,
    /// If true, return all log lines
    #[prost(bool, optional, tag = "5")]
    pub return_all_logs: ::core::option::Option<bool>,
    /// If \[return_all_logs\] is false, return \[num_log_lines\]
    #[prost(uint32, optional, tag = "6")]
    pub num_log_lines: ::core::option::Option<u32>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetServiceLogsResponse {
    /// The service log lines grouped by service UUIDs and ordered in forward direction (oldest log line is the first element)
    #[prost(map = "string, message", tag = "1")]
    pub service_logs_by_service_uuid: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        LogLine,
    >,
    /// A set of service GUIDs requested by the user that were not found in the logs database, could be related that users send
    /// a wrong GUID or a right GUID for a service that has not sent any logs so far
    #[prost(map = "string, bool", tag = "2")]
    pub not_found_service_uuid_set: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        bool,
    >,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogLine {
    #[prost(string, repeated, tag = "1")]
    pub line: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(message, optional, tag = "2")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogLineFilter {
    #[prost(enumeration = "LogLineOperator", tag = "1")]
    pub operator: i32,
    #[prost(string, tag = "2")]
    pub text_pattern: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum EnclaveMode {
    Test = 0,
    Production = 1,
}
impl EnclaveMode {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            EnclaveMode::Test => "TEST",
            EnclaveMode::Production => "PRODUCTION",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "TEST" => Some(Self::Test),
            "PRODUCTION" => Some(Self::Production),
            _ => None,
        }
    }
}
/// ==============================================================================================
///                                             Get Enclaves
/// ==============================================================================================
/// Status of the containers in the enclave
/// NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum valuee uniqueness works
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum EnclaveContainersStatus {
    /// The enclave has been created, but there are no containers inside it
    Empty = 0,
    /// One or more containers are running in the enclave (which may or may not include the API container, depending on if the user was manually stopping/removing containers)
    Running = 1,
    /// There are >= 1 container in the enclave, but they're all stopped
    Stopped = 2,
}
impl EnclaveContainersStatus {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            EnclaveContainersStatus::Empty => "EnclaveContainersStatus_EMPTY",
            EnclaveContainersStatus::Running => "EnclaveContainersStatus_RUNNING",
            EnclaveContainersStatus::Stopped => "EnclaveContainersStatus_STOPPED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "EnclaveContainersStatus_EMPTY" => Some(Self::Empty),
            "EnclaveContainersStatus_RUNNING" => Some(Self::Running),
            "EnclaveContainersStatus_STOPPED" => Some(Self::Stopped),
            _ => None,
        }
    }
}
/// NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum EnclaveApiContainerStatus {
    /// No API container exists in the enclave
    /// This is the only valid value when the enclave containers status is "EMPTY"
    Nonexistent = 0,
    /// An API container exists and is running
    /// NOTE: this does NOT say that the server inside the API container is available, because checking if it's available requires making a call to the API container
    ///   If we have a lot of API containers, we'd be making tons of calls
    Running = 1,
    /// An API container exists, but isn't running
    Stopped = 2,
}
impl EnclaveApiContainerStatus {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            EnclaveApiContainerStatus::Nonexistent => {
                "EnclaveAPIContainerStatus_NONEXISTENT"
            }
            EnclaveApiContainerStatus::Running => "EnclaveAPIContainerStatus_RUNNING",
            EnclaveApiContainerStatus::Stopped => "EnclaveAPIContainerStatus_STOPPED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "EnclaveAPIContainerStatus_NONEXISTENT" => Some(Self::Nonexistent),
            "EnclaveAPIContainerStatus_RUNNING" => Some(Self::Running),
            "EnclaveAPIContainerStatus_STOPPED" => Some(Self::Stopped),
            _ => None,
        }
    }
}
/// The filter operator which can be text or regex type
/// NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum LogLineOperator {
    DoesContainText = 0,
    DoesNotContainText = 1,
    DoesContainMatchRegex = 2,
    DoesNotContainMatchRegex = 3,
}
impl LogLineOperator {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            LogLineOperator::DoesContainText => "LogLineOperator_DOES_CONTAIN_TEXT",
            LogLineOperator::DoesNotContainText => {
                "LogLineOperator_DOES_NOT_CONTAIN_TEXT"
            }
            LogLineOperator::DoesContainMatchRegex => {
                "LogLineOperator_DOES_CONTAIN_MATCH_REGEX"
            }
            LogLineOperator::DoesNotContainMatchRegex => {
                "LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX"
            }
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "LogLineOperator_DOES_CONTAIN_TEXT" => Some(Self::DoesContainText),
            "LogLineOperator_DOES_NOT_CONTAIN_TEXT" => Some(Self::DoesNotContainText),
            "LogLineOperator_DOES_CONTAIN_MATCH_REGEX" => {
                Some(Self::DoesContainMatchRegex)
            }
            "LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX" => {
                Some(Self::DoesNotContainMatchRegex)
            }
            _ => None,
        }
    }
}
/// Generated client implementations.
pub mod engine_service_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct EngineServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl EngineServiceClient<tonic::transport::Channel> {
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
    impl<T> EngineServiceClient<T>
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
        ) -> EngineServiceClient<InterceptedService<T, F>>
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
            EngineServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /// Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
        pub async fn get_engine_info(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetEngineInfoResponse>,
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
                "/engine_api.EngineService/GetEngineInfo",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "GetEngineInfo"));
            self.inner.unary(req, path, codec).await
        }
        /// ==============================================================================================
        ///                                   Enclave Management
        /// ==============================================================================================
        /// Creates a new Kurtosis Enclave
        pub async fn create_enclave(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateEnclaveArgs>,
        ) -> std::result::Result<
            tonic::Response<super::CreateEnclaveResponse>,
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
                "/engine_api.EngineService/CreateEnclave",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "CreateEnclave"));
            self.inner.unary(req, path, codec).await
        }
        /// Returns information about the existing enclaves
        pub async fn get_enclaves(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetEnclavesResponse>,
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
                "/engine_api.EngineService/GetEnclaves",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "GetEnclaves"));
            self.inner.unary(req, path, codec).await
        }
        /// Returns information about all existing & historical enclaves
        pub async fn get_existing_and_historical_enclave_identifiers(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetExistingAndHistoricalEnclaveIdentifiersResponse>,
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
                "/engine_api.EngineService/GetExistingAndHistoricalEnclaveIdentifiers",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "engine_api.EngineService",
                        "GetExistingAndHistoricalEnclaveIdentifiers",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /// Stops all containers in an enclave
        pub async fn stop_enclave(
            &mut self,
            request: impl tonic::IntoRequest<super::StopEnclaveArgs>,
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
                "/engine_api.EngineService/StopEnclave",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "StopEnclave"));
            self.inner.unary(req, path, codec).await
        }
        /// Destroys an enclave, removing all artifacts associated with it
        pub async fn destroy_enclave(
            &mut self,
            request: impl tonic::IntoRequest<super::DestroyEnclaveArgs>,
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
                "/engine_api.EngineService/DestroyEnclave",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "DestroyEnclave"));
            self.inner.unary(req, path, codec).await
        }
        /// Gets rid of old enclaves
        pub async fn clean(
            &mut self,
            request: impl tonic::IntoRequest<super::CleanArgs>,
        ) -> std::result::Result<tonic::Response<super::CleanResponse>, tonic::Status> {
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
                "/engine_api.EngineService/Clean",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "Clean"));
            self.inner.unary(req, path, codec).await
        }
        /// Get service logs
        pub async fn get_service_logs(
            &mut self,
            request: impl tonic::IntoRequest<super::GetServiceLogsArgs>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::GetServiceLogsResponse>>,
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
                "/engine_api.EngineService/GetServiceLogs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("engine_api.EngineService", "GetServiceLogs"));
            self.inner.server_streaming(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod engine_service_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with EngineServiceServer.
    #[async_trait]
    pub trait EngineService: Send + Sync + 'static {
        /// Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
        async fn get_engine_info(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetEngineInfoResponse>,
            tonic::Status,
        >;
        /// ==============================================================================================
        ///                                   Enclave Management
        /// ==============================================================================================
        /// Creates a new Kurtosis Enclave
        async fn create_enclave(
            &self,
            request: tonic::Request<super::CreateEnclaveArgs>,
        ) -> std::result::Result<
            tonic::Response<super::CreateEnclaveResponse>,
            tonic::Status,
        >;
        /// Returns information about the existing enclaves
        async fn get_enclaves(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetEnclavesResponse>,
            tonic::Status,
        >;
        /// Returns information about all existing & historical enclaves
        async fn get_existing_and_historical_enclave_identifiers(
            &self,
            request: tonic::Request<()>,
        ) -> std::result::Result<
            tonic::Response<super::GetExistingAndHistoricalEnclaveIdentifiersResponse>,
            tonic::Status,
        >;
        /// Stops all containers in an enclave
        async fn stop_enclave(
            &self,
            request: tonic::Request<super::StopEnclaveArgs>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status>;
        /// Destroys an enclave, removing all artifacts associated with it
        async fn destroy_enclave(
            &self,
            request: tonic::Request<super::DestroyEnclaveArgs>,
        ) -> std::result::Result<tonic::Response<()>, tonic::Status>;
        /// Gets rid of old enclaves
        async fn clean(
            &self,
            request: tonic::Request<super::CleanArgs>,
        ) -> std::result::Result<tonic::Response<super::CleanResponse>, tonic::Status>;
        /// Server streaming response type for the GetServiceLogs method.
        type GetServiceLogsStream: futures_core::Stream<
                Item = std::result::Result<super::GetServiceLogsResponse, tonic::Status>,
            >
            + Send
            + 'static;
        /// Get service logs
        async fn get_service_logs(
            &self,
            request: tonic::Request<super::GetServiceLogsArgs>,
        ) -> std::result::Result<
            tonic::Response<Self::GetServiceLogsStream>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct EngineServiceServer<T: EngineService> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: EngineService> EngineServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for EngineServiceServer<T>
    where
        T: EngineService,
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
                "/engine_api.EngineService/GetEngineInfo" => {
                    #[allow(non_camel_case_types)]
                    struct GetEngineInfoSvc<T: EngineService>(pub Arc<T>);
                    impl<T: EngineService> tonic::server::UnaryService<()>
                    for GetEngineInfoSvc<T> {
                        type Response = super::GetEngineInfoResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).get_engine_info(request).await
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
                        let method = GetEngineInfoSvc(inner);
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
                "/engine_api.EngineService/CreateEnclave" => {
                    #[allow(non_camel_case_types)]
                    struct CreateEnclaveSvc<T: EngineService>(pub Arc<T>);
                    impl<
                        T: EngineService,
                    > tonic::server::UnaryService<super::CreateEnclaveArgs>
                    for CreateEnclaveSvc<T> {
                        type Response = super::CreateEnclaveResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateEnclaveArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).create_enclave(request).await
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
                        let method = CreateEnclaveSvc(inner);
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
                "/engine_api.EngineService/GetEnclaves" => {
                    #[allow(non_camel_case_types)]
                    struct GetEnclavesSvc<T: EngineService>(pub Arc<T>);
                    impl<T: EngineService> tonic::server::UnaryService<()>
                    for GetEnclavesSvc<T> {
                        type Response = super::GetEnclavesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).get_enclaves(request).await
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
                        let method = GetEnclavesSvc(inner);
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
                "/engine_api.EngineService/GetExistingAndHistoricalEnclaveIdentifiers" => {
                    #[allow(non_camel_case_types)]
                    struct GetExistingAndHistoricalEnclaveIdentifiersSvc<
                        T: EngineService,
                    >(
                        pub Arc<T>,
                    );
                    impl<T: EngineService> tonic::server::UnaryService<()>
                    for GetExistingAndHistoricalEnclaveIdentifiersSvc<T> {
                        type Response = super::GetExistingAndHistoricalEnclaveIdentifiersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner)
                                    .get_existing_and_historical_enclave_identifiers(request)
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
                        let method = GetExistingAndHistoricalEnclaveIdentifiersSvc(
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
                "/engine_api.EngineService/StopEnclave" => {
                    #[allow(non_camel_case_types)]
                    struct StopEnclaveSvc<T: EngineService>(pub Arc<T>);
                    impl<
                        T: EngineService,
                    > tonic::server::UnaryService<super::StopEnclaveArgs>
                    for StopEnclaveSvc<T> {
                        type Response = ();
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::StopEnclaveArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).stop_enclave(request).await
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
                        let method = StopEnclaveSvc(inner);
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
                "/engine_api.EngineService/DestroyEnclave" => {
                    #[allow(non_camel_case_types)]
                    struct DestroyEnclaveSvc<T: EngineService>(pub Arc<T>);
                    impl<
                        T: EngineService,
                    > tonic::server::UnaryService<super::DestroyEnclaveArgs>
                    for DestroyEnclaveSvc<T> {
                        type Response = ();
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DestroyEnclaveArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).destroy_enclave(request).await
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
                        let method = DestroyEnclaveSvc(inner);
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
                "/engine_api.EngineService/Clean" => {
                    #[allow(non_camel_case_types)]
                    struct CleanSvc<T: EngineService>(pub Arc<T>);
                    impl<T: EngineService> tonic::server::UnaryService<super::CleanArgs>
                    for CleanSvc<T> {
                        type Response = super::CleanResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CleanArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move { (*inner).clean(request).await };
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
                        let method = CleanSvc(inner);
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
                "/engine_api.EngineService/GetServiceLogs" => {
                    #[allow(non_camel_case_types)]
                    struct GetServiceLogsSvc<T: EngineService>(pub Arc<T>);
                    impl<
                        T: EngineService,
                    > tonic::server::ServerStreamingService<super::GetServiceLogsArgs>
                    for GetServiceLogsSvc<T> {
                        type Response = super::GetServiceLogsResponse;
                        type ResponseStream = T::GetServiceLogsStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetServiceLogsArgs>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                (*inner).get_service_logs(request).await
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
                        let method = GetServiceLogsSvc(inner);
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
    impl<T: EngineService> Clone for EngineServiceServer<T> {
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
    impl<T: EngineService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(Arc::clone(&self.0))
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: EngineService> tonic::server::NamedService for EngineServiceServer<T> {
        const NAME: &'static str = "engine_api.EngineService";
    }
}
