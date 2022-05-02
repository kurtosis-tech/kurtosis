// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var api_container_service_pb = require('./api_container_service_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

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

function serialize_api_container_api_ExecuteModuleArgs(arg) {
  if (!(arg instanceof api_container_service_pb.ExecuteModuleArgs)) {
    throw new Error('Expected argument of type api_container_api.ExecuteModuleArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ExecuteModuleArgs(buffer_arg) {
  return api_container_service_pb.ExecuteModuleArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_ExecuteModuleResponse(arg) {
  if (!(arg instanceof api_container_service_pb.ExecuteModuleResponse)) {
    throw new Error('Expected argument of type api_container_api.ExecuteModuleResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_ExecuteModuleResponse(buffer_arg) {
  return api_container_service_pb.ExecuteModuleResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetModuleInfoArgs(arg) {
  if (!(arg instanceof api_container_service_pb.GetModuleInfoArgs)) {
    throw new Error('Expected argument of type api_container_api.GetModuleInfoArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetModuleInfoArgs(buffer_arg) {
  return api_container_service_pb.GetModuleInfoArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetModuleInfoResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetModuleInfoResponse)) {
    throw new Error('Expected argument of type api_container_api.GetModuleInfoResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetModuleInfoResponse(buffer_arg) {
  return api_container_service_pb.GetModuleInfoResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetModulesResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetModulesResponse)) {
    throw new Error('Expected argument of type api_container_api.GetModulesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetModulesResponse(buffer_arg) {
  return api_container_service_pb.GetModulesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetServiceInfoArgs(arg) {
  if (!(arg instanceof api_container_service_pb.GetServiceInfoArgs)) {
    throw new Error('Expected argument of type api_container_api.GetServiceInfoArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetServiceInfoArgs(buffer_arg) {
  return api_container_service_pb.GetServiceInfoArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_GetServiceInfoResponse(arg) {
  if (!(arg instanceof api_container_service_pb.GetServiceInfoResponse)) {
    throw new Error('Expected argument of type api_container_api.GetServiceInfoResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_GetServiceInfoResponse(buffer_arg) {
  return api_container_service_pb.GetServiceInfoResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_LoadModuleArgs(arg) {
  if (!(arg instanceof api_container_service_pb.LoadModuleArgs)) {
    throw new Error('Expected argument of type api_container_api.LoadModuleArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_LoadModuleArgs(buffer_arg) {
  return api_container_service_pb.LoadModuleArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_LoadModuleResponse(arg) {
  if (!(arg instanceof api_container_service_pb.LoadModuleResponse)) {
    throw new Error('Expected argument of type api_container_api.LoadModuleResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_LoadModuleResponse(buffer_arg) {
  return api_container_service_pb.LoadModuleResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_RegisterServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RegisterServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.RegisterServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RegisterServiceArgs(buffer_arg) {
  return api_container_service_pb.RegisterServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_RegisterServiceResponse(arg) {
  if (!(arg instanceof api_container_service_pb.RegisterServiceResponse)) {
    throw new Error('Expected argument of type api_container_api.RegisterServiceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RegisterServiceResponse(buffer_arg) {
  return api_container_service_pb.RegisterServiceResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_RepartitionArgs(arg) {
  if (!(arg instanceof api_container_service_pb.RepartitionArgs)) {
    throw new Error('Expected argument of type api_container_api.RepartitionArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_RepartitionArgs(buffer_arg) {
  return api_container_service_pb.RepartitionArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StartServiceArgs(arg) {
  if (!(arg instanceof api_container_service_pb.StartServiceArgs)) {
    throw new Error('Expected argument of type api_container_api.StartServiceArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StartServiceArgs(buffer_arg) {
  return api_container_service_pb.StartServiceArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_api_container_api_StartServiceResponse(arg) {
  if (!(arg instanceof api_container_service_pb.StartServiceResponse)) {
    throw new Error('Expected argument of type api_container_api.StartServiceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_StartServiceResponse(buffer_arg) {
  return api_container_service_pb.StartServiceResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_api_container_api_UnloadModuleArgs(arg) {
  if (!(arg instanceof api_container_service_pb.UnloadModuleArgs)) {
    throw new Error('Expected argument of type api_container_api.UnloadModuleArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_api_container_api_UnloadModuleArgs(buffer_arg) {
  return api_container_service_pb.UnloadModuleArgs.deserializeBinary(new Uint8Array(buffer_arg));
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
  // Starts a module container in the enclave
loadModule: {
    path: '/api_container_api.ApiContainerService/LoadModule',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.LoadModuleArgs,
    responseType: api_container_service_pb.LoadModuleResponse,
    requestSerialize: serialize_api_container_api_LoadModuleArgs,
    requestDeserialize: deserialize_api_container_api_LoadModuleArgs,
    responseSerialize: serialize_api_container_api_LoadModuleResponse,
    responseDeserialize: deserialize_api_container_api_LoadModuleResponse,
  },
  // Stop and remove a module from the enclave
unloadModule: {
    path: '/api_container_api.ApiContainerService/UnloadModule',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.UnloadModuleArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_UnloadModuleArgs,
    requestDeserialize: deserialize_api_container_api_UnloadModuleArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Executes an executable module on the user's behalf
executeModule: {
    path: '/api_container_api.ApiContainerService/ExecuteModule',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.ExecuteModuleArgs,
    responseType: api_container_service_pb.ExecuteModuleResponse,
    requestSerialize: serialize_api_container_api_ExecuteModuleArgs,
    requestDeserialize: deserialize_api_container_api_ExecuteModuleArgs,
    responseSerialize: serialize_api_container_api_ExecuteModuleResponse,
    responseDeserialize: deserialize_api_container_api_ExecuteModuleResponse,
  },
  // Gets information about a loaded module
getModuleInfo: {
    path: '/api_container_api.ApiContainerService/GetModuleInfo',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.GetModuleInfoArgs,
    responseType: api_container_service_pb.GetModuleInfoResponse,
    requestSerialize: serialize_api_container_api_GetModuleInfoArgs,
    requestDeserialize: deserialize_api_container_api_GetModuleInfoArgs,
    responseSerialize: serialize_api_container_api_GetModuleInfoResponse,
    responseDeserialize: deserialize_api_container_api_GetModuleInfoResponse,
  },
  // Registers a service with the API container but doesn't start the container for it
registerService: {
    path: '/api_container_api.ApiContainerService/RegisterService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.RegisterServiceArgs,
    responseType: api_container_service_pb.RegisterServiceResponse,
    requestSerialize: serialize_api_container_api_RegisterServiceArgs,
    requestDeserialize: deserialize_api_container_api_RegisterServiceArgs,
    responseSerialize: serialize_api_container_api_RegisterServiceResponse,
    responseDeserialize: deserialize_api_container_api_RegisterServiceResponse,
  },
  // Starts a previously-registered service by creating a Docker container for it
startService: {
    path: '/api_container_api.ApiContainerService/StartService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.StartServiceArgs,
    responseType: api_container_service_pb.StartServiceResponse,
    requestSerialize: serialize_api_container_api_StartServiceArgs,
    requestDeserialize: deserialize_api_container_api_StartServiceArgs,
    responseSerialize: serialize_api_container_api_StartServiceResponse,
    responseDeserialize: deserialize_api_container_api_StartServiceResponse,
  },
  // Returns relevant information about the service
getServiceInfo: {
    path: '/api_container_api.ApiContainerService/GetServiceInfo',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.GetServiceInfoArgs,
    responseType: api_container_service_pb.GetServiceInfoResponse,
    requestSerialize: serialize_api_container_api_GetServiceInfoArgs,
    requestDeserialize: deserialize_api_container_api_GetServiceInfoArgs,
    responseSerialize: serialize_api_container_api_GetServiceInfoResponse,
    responseDeserialize: deserialize_api_container_api_GetServiceInfoResponse,
  },
  // Instructs the API container to remove the given service
removeService: {
    path: '/api_container_api.ApiContainerService/RemoveService',
    requestStream: false,
    responseStream: false,
    requestType: api_container_service_pb.RemoveServiceArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_api_container_api_RemoveServiceArgs,
    requestDeserialize: deserialize_api_container_api_RemoveServiceArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
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
  // Returns the IDs of the current services in the enclave
getServices: {
    path: '/api_container_api.ApiContainerService/GetServices',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: api_container_service_pb.GetServicesResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_api_container_api_GetServicesResponse,
    responseDeserialize: deserialize_api_container_api_GetServicesResponse,
  },
  // Returns the IDs of the Kurtosis modules that have been loaded into the enclave
getModules: {
    path: '/api_container_api.ApiContainerService/GetModules',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: api_container_service_pb.GetModulesResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_api_container_api_GetModulesResponse,
    responseDeserialize: deserialize_api_container_api_GetModulesResponse,
  },
  // Uploads a files artifact to the Kurtosis File System.
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
};

exports.ApiContainerServiceClient = grpc.makeGenericClientConstructor(ApiContainerServiceService);
