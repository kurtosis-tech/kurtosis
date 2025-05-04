// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var api_container_service_pb = require('./api_container_service_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

function serialize_api_container_api_ConnectServicesArgs(arg) {
  if (!(arg instanceof api_container_service_pb.ConnectServicesArgs)) {
    throw new Error('Expected argument of type api_container_api.ConnectServicesArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ConnectServicesArgs(buffer_arg) {
  return api_container_service_pb.ConnectServicesArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_ConnectServicesResponse(arg) {
  if (!(arg instanceof api_container_service_pb.ConnectServicesResponse)) {
    throw new Error('Expected argument of type api_container_api.ConnectServicesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ConnectServicesResponse(buffer_arg) {
  return api_container_service_pb.ConnectServicesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_CreateSnapshotArgs(arg) {
  if (!(arg instanceof api_container_service_pb.CreateSnapshotArgs)) {
    throw new Error('Expected argument of type api_container_api.CreateSnapshotArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_CreateSnapshotArgs(buffer_arg) {
  return api_container_service_pb.CreateSnapshotArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_DownloadFilesArtifactArgs(arg) {
  if (!(arg instanceof api_container_service_pb.DownloadFilesArtifactArgs)) {
    throw new Error('Expected argument of type api_container_api.DownloadFilesArtifactArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_DownloadFilesArtifactArgs(buffer_arg) {
  return api_container_service_pb.DownloadFilesArtifactArgs.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_GetStarlarkRunResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetStarlarkRunResponse)) {
    throw new Error('Expected argument of type api_container_api.GetStarlarkRunResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetStarlarkRunResponse(buffer_arg) {
  return api_container_service_pb.GetStarlarkRunResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_InspectFilesArtifactContentsRequest(arg) {
  if (!(arg instanceof api_container_service_pb.InspectFilesArtifactContentsRequest)) {
    throw new Error('Expected argument of type api_container_api.InspectFilesArtifactContentsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_InspectFilesArtifactContentsRequest(buffer_arg) {
  return api_container_service_pb.InspectFilesArtifactContentsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_InspectFilesArtifactContentsResponse(arg) {
  if (!(arg instanceof api_container_service_pb.InspectFilesArtifactContentsResponse)) {
    throw new Error('Expected argument of type api_container_api.InspectFilesArtifactContentsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_InspectFilesArtifactContentsResponse(buffer_arg) {
  return api_container_service_pb.InspectFilesArtifactContentsResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_PlanYaml(arg) {
  if (!(arg instanceof api_container_service_pb.PlanYaml)) {
    throw new Error('Expected argument of type api_container_api.PlanYaml');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_PlanYaml(buffer_arg) {
  return api_container_service_pb.PlanYaml.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_StarlarkPackagePlanYamlArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StarlarkPackagePlanYamlArgs)) {
    throw new Error('Expected argument of type api_container_api.StarlarkPackagePlanYamlArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StarlarkPackagePlanYamlArgs(buffer_arg) {
  return api_container_service_pb.StarlarkPackagePlanYamlArgs.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_StarlarkScriptPlanYamlArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StarlarkScriptPlanYamlArgs)) {
    throw new Error('Expected argument of type api_container_api.StarlarkScriptPlanYamlArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StarlarkScriptPlanYamlArgs(buffer_arg) {
  return api_container_service_pb.StarlarkScriptPlanYamlArgs.deserializeBinary(new Uint8Array(buffer_arg));
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
uploadFilesArtifact: {
    path: '/api_container_api.ApiContainerService/UploadFilesArtifact',
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
downloadFilesArtifact: {
    path: '/api_container_api.ApiContainerService/DownloadFilesArtifact',
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
  inspectFilesArtifactContents: {
    path: '/api_container_api.ApiContainerService/InspectFilesArtifactContents',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.InspectFilesArtifactContentsRequest,
    responseType: api_container_service_pb.InspectFilesArtifactContentsResponse,
    requestSerialize: serialize_api_container_api_InspectFilesArtifactContentsRequest,
    requestDeserialize: deserialize_api_container_api_InspectFilesArtifactContentsRequest,
    responseSerialize: serialize_api_container_api_InspectFilesArtifactContentsResponse,
    responseDeserialize: deserialize_api_container_api_InspectFilesArtifactContentsResponse,
  },
  // User services port forwarding
connectServices: {
    path: '/api_container_api.ApiContainerService/ConnectServices',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.ConnectServicesArgs,
    responseType: api_container_service_pb.ConnectServicesResponse,
    requestSerialize: serialize_api_container_api_ConnectServicesArgs,
    requestDeserialize: deserialize_api_container_api_ConnectServicesArgs,
    responseSerialize: serialize_api_container_api_ConnectServicesResponse,
    responseDeserialize: deserialize_api_container_api_ConnectServicesResponse,
  },
  // Get last Starlark run
getStarlarkRun: {
    path: '/api_container_api.ApiContainerService/GetStarlarkRun',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: api_container_service_pb.GetStarlarkRunResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_api_container_api_GetStarlarkRunResponse,
    responseDeserialize: deserialize_api_container_api_GetStarlarkRunResponse,
  },
  // Gets yaml representing the plan the script will execute in an enclave
getStarlarkScriptPlanYaml: {
    path: '/api_container_api.ApiContainerService/GetStarlarkScriptPlanYaml',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StarlarkScriptPlanYamlArgs,
    responseType: api_container_service_pb.PlanYaml,
    requestSerialize: serialize_api_container_api_StarlarkScriptPlanYamlArgs,
    requestDeserialize: deserialize_api_container_api_StarlarkScriptPlanYamlArgs,
    responseSerialize: serialize_api_container_api_PlanYaml,
    responseDeserialize: deserialize_api_container_api_PlanYaml,
  },
  // Gets yaml representing the plan the package will execute in an enclave
getStarlarkPackagePlanYaml: {
    path: '/api_container_api.ApiContainerService/GetStarlarkPackagePlanYaml',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StarlarkPackagePlanYamlArgs,
    responseType: api_container_service_pb.PlanYaml,
    requestSerialize: serialize_api_container_api_StarlarkPackagePlanYamlArgs,
    requestDeserialize: deserialize_api_container_api_StarlarkPackagePlanYamlArgs,
    responseSerialize: serialize_api_container_api_PlanYaml,
    responseDeserialize: deserialize_api_container_api_PlanYaml,
  },
  createSnapshot: {
    path: '/api_container_api.ApiContainerService/CreateSnapshot',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.CreateSnapshotArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_CreateSnapshotArgs,
    requestDeserialize: deserialize_api_container_api_CreateSnapshotArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.ApiContainerServiceClient = grpc.makeGenericClientConstructor(ApiContainerServiceService);
