// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var api_container_service_pb = require('./api_container_service_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

function serialize_api_container_api_DownloadFilesArtifactArgs(arg) {
  if (!(arg instanceof api_container_service_pb.DownloadFilesArtifactArgs)) {
    throw new Error('Expected argument of type api_container_api.DownloadFilesArtifactArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_DownloadFilesArtifactArgs(buffer_arg) {
  return api_container_service_pb.DownloadFilesArtifactArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_DownloadFilesArtifactResponse(arg) {
  if (!(arg instanceof api_container_service_pb.DownloadFilesArtifactResponse)) {
    throw new Error('Expected argument of type api_container_api.DownloadFilesArtifactResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_DownloadFilesArtifactResponse(buffer_arg) {
  return api_container_service_pb.DownloadFilesArtifactResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_ExecCommandArgs(arg) {
  if (!(arg instanceof api_container_service_pb.ExecCommandArgs)) {
    throw new Error('Expected argument of type api_container_api.ExecCommandArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ExecCommandArgs(buffer_arg) {
  return api_container_service_pb.ExecCommandArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_ExecCommandResponse(arg) {
  if (!(arg instanceof api_container_service_pb.ExecCommandResponse)) {
    throw new Error('Expected argument of type api_container_api.ExecCommandResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ExecCommandResponse(buffer_arg) {
  return api_container_service_pb.ExecCommandResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetExistingAndHistoricalServiceIdentifiersResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse)) {
    throw new Error('Expected argument of type api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetExistingAndHistoricalServiceIdentifiersResponse(buffer_arg) {
  return api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetServicesArgs(arg) {
  if (!(arg instanceof api_container_service_pb.GetServicesArgs)) {
    throw new Error('Expected argument of type api_container_api.GetServicesArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetServicesArgs(buffer_arg) {
  return api_container_service_pb.GetServicesArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetServicesResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetServicesResponse)) {
    throw new Error('Expected argument of type api_container_api.GetServicesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetServicesResponse(buffer_arg) {
  return api_container_service_pb.GetServicesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_ListFilesArtifactNamesAndUuidsResponse(arg) {
  if (!(arg instanceof api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse)) {
    throw new Error('Expected argument of type api_container_api.ListFilesArtifactNamesAndUuidsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ListFilesArtifactNamesAndUuidsResponse(buffer_arg) {
  return api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_PauseServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.PauseServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.PauseServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_PauseServiceArgs(buffer_arg) {
  return api_container_service_pb.PauseServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RemoveServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RemoveServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.RemoveServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RemoveServiceArgs(buffer_arg) {
  return api_container_service_pb.RemoveServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RemoveServiceResponse(arg) {
  if (!(arg instanceof api_container_service_pb.RemoveServiceResponse)) {
    throw new Error('Expected argument of type api_container_api.RemoveServiceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RemoveServiceResponse(buffer_arg) {
  return api_container_service_pb.RemoveServiceResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RenderTemplatesToFilesArtifactArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RenderTemplatesToFilesArtifactArgs)) {
    throw new Error('Expected argument of type api_container_api.RenderTemplatesToFilesArtifactArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RenderTemplatesToFilesArtifactArgs(buffer_arg) {
  return api_container_service_pb.RenderTemplatesToFilesArtifactArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RenderTemplatesToFilesArtifactResponse(arg) {
  if (!(arg instanceof api_container_service_pb.RenderTemplatesToFilesArtifactResponse)) {
    throw new Error('Expected argument of type api_container_api.RenderTemplatesToFilesArtifactResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RenderTemplatesToFilesArtifactResponse(buffer_arg) {
  return api_container_service_pb.RenderTemplatesToFilesArtifactResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RepartitionArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RepartitionArgs)) {
    throw new Error('Expected argument of type api_container_api.RepartitionArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RepartitionArgs(buffer_arg) {
  return api_container_service_pb.RepartitionArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RunStarlarkPackageArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RunStarlarkPackageArgs)) {
    throw new Error('Expected argument of type api_container_api.RunStarlarkPackageArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RunStarlarkPackageArgs(buffer_arg) {
  return api_container_service_pb.RunStarlarkPackageArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RunStarlarkScriptArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RunStarlarkScriptArgs)) {
    throw new Error('Expected argument of type api_container_api.RunStarlarkScriptArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RunStarlarkScriptArgs(buffer_arg) {
  return api_container_service_pb.RunStarlarkScriptArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StarlarkRunResponseLine(arg) {
  if (!(arg instanceof api_container_service_pb.StarlarkRunResponseLine)) {
    throw new Error('Expected argument of type api_container_api.StarlarkRunResponseLine');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StarlarkRunResponseLine(buffer_arg) {
  return api_container_service_pb.StarlarkRunResponseLine.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StartServicesArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StartServicesArgs)) {
    throw new Error('Expected argument of type api_container_api.StartServicesArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StartServicesArgs(buffer_arg) {
  return api_container_service_pb.StartServicesArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StartServicesResponse(arg) {
  if (!(arg instanceof api_container_service_pb.StartServicesResponse)) {
    throw new Error('Expected argument of type api_container_api.StartServicesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StartServicesResponse(buffer_arg) {
  return api_container_service_pb.StartServicesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StoreFilesArtifactFromServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StoreFilesArtifactFromServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.StoreFilesArtifactFromServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StoreFilesArtifactFromServiceArgs(buffer_arg) {
  return api_container_service_pb.StoreFilesArtifactFromServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StoreFilesArtifactFromServiceResponse(arg) {
  if (!(arg instanceof api_container_service_pb.StoreFilesArtifactFromServiceResponse)) {
    throw new Error('Expected argument of type api_container_api.StoreFilesArtifactFromServiceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StoreFilesArtifactFromServiceResponse(buffer_arg) {
  return api_container_service_pb.StoreFilesArtifactFromServiceResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StoreWebFilesArtifactArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StoreWebFilesArtifactArgs)) {
    throw new Error('Expected argument of type api_container_api.StoreWebFilesArtifactArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StoreWebFilesArtifactArgs(buffer_arg) {
  return api_container_service_pb.StoreWebFilesArtifactArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StoreWebFilesArtifactResponse(arg) {
  if (!(arg instanceof api_container_service_pb.StoreWebFilesArtifactResponse)) {
    throw new Error('Expected argument of type api_container_api.StoreWebFilesArtifactResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StoreWebFilesArtifactResponse(buffer_arg) {
  return api_container_service_pb.StoreWebFilesArtifactResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StreamedDataChunk(arg) {
  if (!(arg instanceof api_container_service_pb.StreamedDataChunk)) {
    throw new Error('Expected argument of type api_container_api.StreamedDataChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StreamedDataChunk(buffer_arg) {
  return api_container_service_pb.StreamedDataChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_UnpauseServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.UnpauseServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.UnpauseServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_UnpauseServiceArgs(buffer_arg) {
  return api_container_service_pb.UnpauseServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_UploadFilesArtifactArgs(arg) {
  if (!(arg instanceof api_container_service_pb.UploadFilesArtifactArgs)) {
    throw new Error('Expected argument of type api_container_api.UploadFilesArtifactArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_UploadFilesArtifactArgs(buffer_arg) {
  return api_container_service_pb.UploadFilesArtifactArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_UploadFilesArtifactResponse(arg) {
  if (!(arg instanceof api_container_service_pb.UploadFilesArtifactResponse)) {
    throw new Error('Expected argument of type api_container_api.UploadFilesArtifactResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_UploadFilesArtifactResponse(buffer_arg) {
  return api_container_service_pb.UploadFilesArtifactResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_WaitForHttpGetEndpointAvailabilityArgs(arg) {
  if (!(arg instanceof api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs)) {
    throw new Error('Expected argument of type api_container_api.WaitForHttpGetEndpointAvailabilityArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_WaitForHttpGetEndpointAvailabilityArgs(buffer_arg) {
  return api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_WaitForHttpPostEndpointAvailabilityArgs(arg) {
  if (!(arg instanceof api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs)) {
    throw new Error('Expected argument of type api_container_api.WaitForHttpPostEndpointAvailabilityArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_WaitForHttpPostEndpointAvailabilityArgs(buffer_arg) {
  return api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}


var ApiContainerServiceService = exports.ApiContainerServiceService = {
  // Executes a Starlark script on the user's behalf
runStarlarkScript: {
    path: '/api_container_api.ApiContainerService/RunStarlarkScript',
    requestStream: false,
    responseStream: true,
    requestType: api_container_service_pb.RunStarlarkScriptArgs,
    responseType: api_container_service_pb.StarlarkRunResponseLine,
    requestSerialize: serialize_api_container_api_RunStarlarkScriptArgs,
    requestDeserialize: deserialize_api_container_api_RunStarlarkScriptArgs,
    responseSerialize: serialize_api_container_api_StarlarkRunResponseLine,
    responseDeserialize: deserialize_api_container_api_StarlarkRunResponseLine,
  },
  // Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
uploadStarlarkPackage: {
    path: '/api_container_api.ApiContainerService/UploadStarlarkPackage',
    requestStream: true,
    responseStream: false,
    requestType: api_container_service_pb.StreamedDataChunk,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_StreamedDataChunk,
    requestDeserialize: deserialize_api_container_api_StreamedDataChunk,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Executes a Starlark script on the user's behalf
runStarlarkPackage: {
    path: '/api_container_api.ApiContainerService/RunStarlarkPackage',
    requestStream: false,
    responseStream: true,
    requestType: api_container_service_pb.RunStarlarkPackageArgs,
    responseType: api_container_service_pb.StarlarkRunResponseLine,
    requestSerialize: serialize_api_container_api_RunStarlarkPackageArgs,
    requestDeserialize: deserialize_api_container_api_RunStarlarkPackageArgs,
    responseSerialize: serialize_api_container_api_StarlarkRunResponseLine,
    responseDeserialize: deserialize_api_container_api_StarlarkRunResponseLine,
  },
  // Start services by creating containers for them
startServices: {
    path: '/api_container_api.ApiContainerService/StartServices',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StartServicesArgs,
    responseType: api_container_service_pb.StartServicesResponse,
    requestSerialize: serialize_api_container_api_StartServicesArgs,
    requestDeserialize: deserialize_api_container_api_StartServicesArgs,
    responseSerialize: serialize_api_container_api_StartServicesResponse,
    responseDeserialize: deserialize_api_container_api_StartServicesResponse,
  },
  // Returns the IDs of the current services in the enclave
getServices: {
    path: '/api_container_api.ApiContainerService/GetServices',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.GetServicesArgs,
    responseType: api_container_service_pb.GetServicesResponse,
    requestSerialize: serialize_api_container_api_GetServicesArgs,
    requestDeserialize: deserialize_api_container_api_GetServicesArgs,
    responseSerialize: serialize_api_container_api_GetServicesResponse,
    responseDeserialize: deserialize_api_container_api_GetServicesResponse,
  },
  // Returns information about all existing & historical services
getExistingAndHistoricalServiceIdentifiers: {
    path: '/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_api_container_api_GetExistingAndHistoricalServiceIdentifiersResponse,
    responseDeserialize: deserialize_api_container_api_GetExistingAndHistoricalServiceIdentifiersResponse,
  },
  // Instructs the API container to remove the given service
removeService: {
    path: '/api_container_api.ApiContainerService/RemoveService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.RemoveServiceArgs,
    responseType: api_container_service_pb.RemoveServiceResponse,
    requestSerialize: serialize_api_container_api_RemoveServiceArgs,
    requestDeserialize: deserialize_api_container_api_RemoveServiceArgs,
    responseSerialize: serialize_api_container_api_RemoveServiceResponse,
    responseDeserialize: deserialize_api_container_api_RemoveServiceResponse,
  },
  // Instructs the API container to repartition the enclave
repartition: {
    path: '/api_container_api.ApiContainerService/Repartition',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.RepartitionArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_RepartitionArgs,
    requestDeserialize: deserialize_api_container_api_RepartitionArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Executes the given command inside a running container
execCommand: {
    path: '/api_container_api.ApiContainerService/ExecCommand',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.ExecCommandArgs,
    responseType: api_container_service_pb.ExecCommandResponse,
    requestSerialize: serialize_api_container_api_ExecCommandArgs,
    requestDeserialize: deserialize_api_container_api_ExecCommandArgs,
    responseSerialize: serialize_api_container_api_ExecCommandResponse,
    responseDeserialize: deserialize_api_container_api_ExecCommandResponse,
  },
  // Pauses all processes running in the service container
pauseService: {
    path: '/api_container_api.ApiContainerService/PauseService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.PauseServiceArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_PauseServiceArgs,
    requestDeserialize: deserialize_api_container_api_PauseServiceArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Unpauses all paused processes running in the service container
unpauseService: {
    path: '/api_container_api.ApiContainerService/UnpauseService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.UnpauseServiceArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_UnpauseServiceArgs,
    requestDeserialize: deserialize_api_container_api_UnpauseServiceArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
waitForHttpGetEndpointAvailability: {
    path: '/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_WaitForHttpGetEndpointAvailabilityArgs,
    requestDeserialize: deserialize_api_container_api_WaitForHttpGetEndpointAvailabilityArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
waitForHttpPostEndpointAvailability: {
    path: '/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_WaitForHttpPostEndpointAvailabilityArgs,
    requestDeserialize: deserialize_api_container_api_WaitForHttpPostEndpointAvailabilityArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Uploads a files artifact to the Kurtosis File System
// Deprecated: please use UploadFilesArtifactV2 to stream the data and not be blocked by the 4MB limit
uploadFilesArtifact: {
    path: '/api_container_api.ApiContainerService/UploadFilesArtifact',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.UploadFilesArtifactArgs,
    responseType: api_container_service_pb.UploadFilesArtifactResponse,
    requestSerialize: serialize_api_container_api_UploadFilesArtifactArgs,
    requestDeserialize: deserialize_api_container_api_UploadFilesArtifactArgs,
    responseSerialize: serialize_api_container_api_UploadFilesArtifactResponse,
    responseDeserialize: deserialize_api_container_api_UploadFilesArtifactResponse,
  },
  // Uploads a files artifact to the Kurtosis File System
// Can be deprecated once we do not use it anymore. For now, it is still used in the TS SDK as grp-file-transfer
// library is only implemented in Go
uploadFilesArtifactV2: {
    path: '/api_container_api.ApiContainerService/UploadFilesArtifactV2',
    requestStream: true,
    responseStream: false,
    requestType: api_container_service_pb.StreamedDataChunk,
    responseType: api_container_service_pb.UploadFilesArtifactResponse,
    requestSerialize: serialize_api_container_api_StreamedDataChunk,
    requestDeserialize: deserialize_api_container_api_StreamedDataChunk,
    responseSerialize: serialize_api_container_api_UploadFilesArtifactResponse,
    responseDeserialize: deserialize_api_container_api_UploadFilesArtifactResponse,
  },
  // Downloads a files artifact from the Kurtosis File System
// Deprecated: Use DownloadFilesArtifactV2 to stream the data and not be limited by GRPC 4MB limit
downloadFilesArtifact: {
    path: '/api_container_api.ApiContainerService/DownloadFilesArtifact',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.DownloadFilesArtifactArgs,
    responseType: api_container_service_pb.DownloadFilesArtifactResponse,
    requestSerialize: serialize_api_container_api_DownloadFilesArtifactArgs,
    requestDeserialize: deserialize_api_container_api_DownloadFilesArtifactArgs,
    responseSerialize: serialize_api_container_api_DownloadFilesArtifactResponse,
    responseDeserialize: deserialize_api_container_api_DownloadFilesArtifactResponse,
  },
  // Downloads a files artifact from the Kurtosis File System
downloadFilesArtifactV2: {
    path: '/api_container_api.ApiContainerService/DownloadFilesArtifactV2',
    requestStream: false,
    responseStream: true,
    requestType: api_container_service_pb.DownloadFilesArtifactArgs,
    responseType: api_container_service_pb.StreamedDataChunk,
    requestSerialize: serialize_api_container_api_DownloadFilesArtifactArgs,
    requestDeserialize: deserialize_api_container_api_DownloadFilesArtifactArgs,
    responseSerialize: serialize_api_container_api_StreamedDataChunk,
    responseDeserialize: deserialize_api_container_api_StreamedDataChunk,
  },
  // Tells the API container to download a files artifact from the web to the Kurtosis File System
storeWebFilesArtifact: {
    path: '/api_container_api.ApiContainerService/StoreWebFilesArtifact',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StoreWebFilesArtifactArgs,
    responseType: api_container_service_pb.StoreWebFilesArtifactResponse,
    requestSerialize: serialize_api_container_api_StoreWebFilesArtifactArgs,
    requestDeserialize: deserialize_api_container_api_StoreWebFilesArtifactArgs,
    responseSerialize: serialize_api_container_api_StoreWebFilesArtifactResponse,
    responseDeserialize: deserialize_api_container_api_StoreWebFilesArtifactResponse,
  },
  // Tells the API container to copy a files artifact from a service to the Kurtosis File System
storeFilesArtifactFromService: {
    path: '/api_container_api.ApiContainerService/StoreFilesArtifactFromService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StoreFilesArtifactFromServiceArgs,
    responseType: api_container_service_pb.StoreFilesArtifactFromServiceResponse,
    requestSerialize: serialize_api_container_api_StoreFilesArtifactFromServiceArgs,
    requestDeserialize: deserialize_api_container_api_StoreFilesArtifactFromServiceArgs,
    responseSerialize: serialize_api_container_api_StoreFilesArtifactFromServiceResponse,
    responseDeserialize: deserialize_api_container_api_StoreFilesArtifactFromServiceResponse,
  },
  // Renders the templates and their data to a files artifact in the Kurtosis File System
renderTemplatesToFilesArtifact: {
    path: '/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.RenderTemplatesToFilesArtifactArgs,
    responseType: api_container_service_pb.RenderTemplatesToFilesArtifactResponse,
    requestSerialize: serialize_api_container_api_RenderTemplatesToFilesArtifactArgs,
    requestDeserialize: deserialize_api_container_api_RenderTemplatesToFilesArtifactArgs,
    responseSerialize: serialize_api_container_api_RenderTemplatesToFilesArtifactResponse,
    responseDeserialize: deserialize_api_container_api_RenderTemplatesToFilesArtifactResponse,
  },
  listFilesArtifactNamesAndUuids: {
    path: '/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_api_container_api_ListFilesArtifactNamesAndUuidsResponse,
    responseDeserialize: deserialize_api_container_api_ListFilesArtifactNamesAndUuidsResponse,
  },
};

exports.ApiContainerServiceClient = grpc.makeGenericClientConstructor(ApiContainerServiceService);
