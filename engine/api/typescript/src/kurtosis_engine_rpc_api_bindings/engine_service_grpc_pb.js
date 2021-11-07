// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var engine_service_pb = require('./engine_service_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

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
};

exports.EngineServiceClient = grpc.makeGenericClientConstructor(EngineServiceService);
