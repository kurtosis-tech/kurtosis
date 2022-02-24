// Own Version
export { KURTOSIS_ENGINE_VERSION } from "./kurtosis_engine_version/kurtosis_engine_version";

export { KurtosisContext, DEFAULT_GRPC_PROXY_ENGINE_SERVER_PORT_NUM, DEFAULT_GRPC_ENGINE_SERVER_PORT_NUM } from "./lib/kurtosis_context/kurtosis_context";

export { EnclaveAPIContainerHostMachineInfo } from "./kurtosis_engine_rpc_api_bindings/engine_service_pb"

// RPC API binding
//TODO: REMOVE THIS LINE AFTER GRPC WEB IS FULLY IMPLEMENTED
export { EngineServiceClient } from "./kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb";

export { EngineServiceClient as EngineServiceClientWeb } from "./kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb";
export { EngineServiceClient as EngineServiceClientNode } from "./kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb";
