// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var engine_service_pb = require('./engine_service_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_engine_api_CleanArgs(arg) {
  if (!(arg instanceof engine_service_pb.CleanArgs)) {
    throw new Error('Expected argument of type engine_api.CleanArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_CleanArgs(buffer_arg) {
  return engine_service_pb.CleanArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_CleanResponse(arg) {
  if (!(arg instanceof engine_service_pb.CleanResponse)) {
    throw new Error('Expected argument of type engine_api.CleanResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_CleanResponse(buffer_arg) {
  return engine_service_pb.CleanResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_CreateEnclaveArgs(arg) {
  if (!(arg instanceof engine_service_pb.CreateEnclaveArgs)) {
    throw new Error('Expected argument of type engine_api.CreateEnclaveArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_CreateEnclaveArgs(buffer_arg) {
  return engine_service_pb.CreateEnclaveArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_CreateEnclaveResponse(arg) {
  if (!(arg instanceof engine_service_pb.CreateEnclaveResponse)) {
    throw new Error('Expected argument of type engine_api.CreateEnclaveResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_CreateEnclaveResponse(buffer_arg) {
  return engine_service_pb.CreateEnclaveResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_DestroyEnclaveArgs(arg) {
  if (!(arg instanceof engine_service_pb.DestroyEnclaveArgs)) {
    throw new Error('Expected argument of type engine_api.DestroyEnclaveArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_DestroyEnclaveArgs(buffer_arg) {
  return engine_service_pb.DestroyEnclaveArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_GetEnclavesResponse(arg) {
  if (!(arg instanceof engine_service_pb.GetEnclavesResponse)) {
    throw new Error('Expected argument of type engine_api.GetEnclavesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_GetEnclavesResponse(buffer_arg) {
  return engine_service_pb.GetEnclavesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_GetEngineInfoResponse(arg) {
  if (!(arg instanceof engine_service_pb.GetEngineInfoResponse)) {
    throw new Error('Expected argument of type engine_api.GetEngineInfoResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_GetEngineInfoResponse(buffer_arg) {
  return engine_service_pb.GetEngineInfoResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_GetExistingAndHistoricalEnclaveIdentifiersResponse(arg) {
  if (!(arg instanceof engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse)) {
    throw new Error('Expected argument of type engine_api.GetExistingAndHistoricalEnclaveIdentifiersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_GetExistingAndHistoricalEnclaveIdentifiersResponse(buffer_arg) {
  return engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_GetServiceLogsArgs(arg) {
  if (!(arg instanceof engine_service_pb.GetServiceLogsArgs)) {
    throw new Error('Expected argument of type engine_api.GetServiceLogsArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_GetServiceLogsArgs(buffer_arg) {
  return engine_service_pb.GetServiceLogsArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_GetServiceLogsResponse(arg) {
  if (!(arg instanceof engine_service_pb.GetServiceLogsResponse)) {
    throw new Error('Expected argument of type engine_api.GetServiceLogsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_GetServiceLogsResponse(buffer_arg) {
  return engine_service_pb.GetServiceLogsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_engine_api_StopEnclaveArgs(arg) {
  if (!(arg instanceof engine_service_pb.StopEnclaveArgs)) {
    throw new Error('Expected argument of type engine_api.StopEnclaveArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_engine_api_StopEnclaveArgs(buffer_arg) {
  return engine_service_pb.StopEnclaveArgs.deserializeBinary(new Uint8Array(buffer_arg));
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


var EngineServiceService = exports.EngineServiceService = {
  // Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
getEngineInfo: {
    path: '/engine_api.EngineService/GetEngineInfo',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: engine_service_pb.GetEngineInfoResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_engine_api_GetEngineInfoResponse,
    responseDeserialize: deserialize_engine_api_GetEngineInfoResponse,
  },
  // ==============================================================================================
//                                   Enclave Management
// ==============================================================================================
// Creates a new Kurtosis Enclave
createEnclave: {
    path: '/engine_api.EngineService/CreateEnclave',
    requestStream: false,
    responseStream: false,
    requestType: engine_service_pb.CreateEnclaveArgs,
    responseType: engine_service_pb.CreateEnclaveResponse,
    requestSerialize: serialize_engine_api_CreateEnclaveArgs,
    requestDeserialize: deserialize_engine_api_CreateEnclaveArgs,
    responseSerialize: serialize_engine_api_CreateEnclaveResponse,
    responseDeserialize: deserialize_engine_api_CreateEnclaveResponse,
  },
  // Returns information about the existing enclaves
getEnclaves: {
    path: '/engine_api.EngineService/GetEnclaves',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: engine_service_pb.GetEnclavesResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_engine_api_GetEnclavesResponse,
    responseDeserialize: deserialize_engine_api_GetEnclavesResponse,
  },
  // Returns information about all existing & historical enclaves
getExistingAndHistoricalEnclaveIdentifiers: {
    path: '/engine_api.EngineService/GetExistingAndHistoricalEnclaveIdentifiers',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_engine_api_GetExistingAndHistoricalEnclaveIdentifiersResponse,
    responseDeserialize: deserialize_engine_api_GetExistingAndHistoricalEnclaveIdentifiersResponse,
  },
  // Stops all containers in an enclave
stopEnclave: {
    path: '/engine_api.EngineService/StopEnclave',
    requestStream: false,
    responseStream: false,
    requestType: engine_service_pb.StopEnclaveArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_engine_api_StopEnclaveArgs,
    requestDeserialize: deserialize_engine_api_StopEnclaveArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Destroys an enclave, removing all artifacts associated with it
destroyEnclave: {
    path: '/engine_api.EngineService/DestroyEnclave',
    requestStream: false,
    responseStream: false,
    requestType: engine_service_pb.DestroyEnclaveArgs,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_engine_api_DestroyEnclaveArgs,
    requestDeserialize: deserialize_engine_api_DestroyEnclaveArgs,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Gets rid of old enclaves
clean: {
    path: '/engine_api.EngineService/Clean',
    requestStream: false,
    responseStream: false,
    requestType: engine_service_pb.CleanArgs,
    responseType: engine_service_pb.CleanResponse,
    requestSerialize: serialize_engine_api_CleanArgs,
    requestDeserialize: deserialize_engine_api_CleanArgs,
    responseSerialize: serialize_engine_api_CleanResponse,
    responseDeserialize: deserialize_engine_api_CleanResponse,
  },
  // Get service logs
getServiceLogs: {
    path: '/engine_api.EngineService/GetServiceLogs',
    requestStream: false,
    responseStream: true,
    requestType: engine_service_pb.GetServiceLogsArgs,
    responseType: engine_service_pb.GetServiceLogsResponse,
    requestSerialize: serialize_engine_api_GetServiceLogsArgs,
    requestDeserialize: deserialize_engine_api_GetServiceLogsArgs,
    responseSerialize: serialize_engine_api_GetServiceLogsResponse,
    responseDeserialize: deserialize_engine_api_GetServiceLogsResponse,
  },
};

exports.EngineServiceClient = grpc.makeGenericClientConstructor(EngineServiceService);
