// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var portal_daemon_service_pb = require('./portal_daemon_service_pb.js');

function serialize_portal_daemon_api_CreateUserServicePortForwardArgs(arg) {
  if (!(arg instanceof portal_daemon_service_pb.CreateUserServicePortForwardArgs)) {
    throw new Error('Expected argument of type portal_daemon_api.CreateUserServicePortForwardArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_CreateUserServicePortForwardArgs(buffer_arg) {
  return portal_daemon_service_pb.CreateUserServicePortForwardArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_CreateUserServicePortForwardResponse(arg) {
  if (!(arg instanceof portal_daemon_service_pb.CreateUserServicePortForwardResponse)) {
    throw new Error('Expected argument of type portal_daemon_api.CreateUserServicePortForwardResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_CreateUserServicePortForwardResponse(buffer_arg) {
  return portal_daemon_service_pb.CreateUserServicePortForwardResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_EnclaveServicePortId(arg) {
  if (!(arg instanceof portal_daemon_service_pb.EnclaveServicePortId)) {
    throw new Error('Expected argument of type portal_daemon_api.EnclaveServicePortId');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_EnclaveServicePortId(buffer_arg) {
  return portal_daemon_service_pb.EnclaveServicePortId.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_PortalPing(arg) {
  if (!(arg instanceof portal_daemon_service_pb.PortalPing)) {
    throw new Error('Expected argument of type portal_daemon_api.PortalPing');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_PortalPing(buffer_arg) {
  return portal_daemon_service_pb.PortalPing.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_PortalPong(arg) {
  if (!(arg instanceof portal_daemon_service_pb.PortalPong)) {
    throw new Error('Expected argument of type portal_daemon_api.PortalPong');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_PortalPong(buffer_arg) {
  return portal_daemon_service_pb.PortalPong.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_RemoveUserServicePortForwardResponse(arg) {
  if (!(arg instanceof portal_daemon_service_pb.RemoveUserServicePortForwardResponse)) {
    throw new Error('Expected argument of type portal_daemon_api.RemoveUserServicePortForwardResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_RemoveUserServicePortForwardResponse(buffer_arg) {
  return portal_daemon_service_pb.RemoveUserServicePortForwardResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var KurtosisPortalDaemonService = exports.KurtosisPortalDaemonService = {
  // To check availability
ping: {
    path: '/portal_daemon_api.KurtosisPortalDaemon/Ping',
    requestStream: false,
    responseStream: false,
    requestType: portal_daemon_service_pb.PortalPing,
    responseType: portal_daemon_service_pb.PortalPong,
    requestSerialize: serialize_portal_daemon_api_PortalPing,
    requestDeserialize: deserialize_portal_daemon_api_PortalPing,
    responseSerialize: serialize_portal_daemon_api_PortalPong,
    responseDeserialize: deserialize_portal_daemon_api_PortalPong,
  },
  createUserServicePortForward: {
    path: '/portal_daemon_api.KurtosisPortalDaemon/CreateUserServicePortForward',
    requestStream: false,
    responseStream: false,
    requestType: portal_daemon_service_pb.CreateUserServicePortForwardArgs,
    responseType: portal_daemon_service_pb.CreateUserServicePortForwardResponse,
    requestSerialize: serialize_portal_daemon_api_CreateUserServicePortForwardArgs,
    requestDeserialize: deserialize_portal_daemon_api_CreateUserServicePortForwardArgs,
    responseSerialize: serialize_portal_daemon_api_CreateUserServicePortForwardResponse,
    responseDeserialize: deserialize_portal_daemon_api_CreateUserServicePortForwardResponse,
  },
  removeUserServicePortForward: {
    path: '/portal_daemon_api.KurtosisPortalDaemon/RemoveUserServicePortForward',
    requestStream: false,
    responseStream: false,
    requestType: portal_daemon_service_pb.EnclaveServicePortId,
    responseType: portal_daemon_service_pb.RemoveUserServicePortForwardResponse,
    requestSerialize: serialize_portal_daemon_api_EnclaveServicePortId,
    requestDeserialize: deserialize_portal_daemon_api_EnclaveServicePortId,
    responseSerialize: serialize_portal_daemon_api_RemoveUserServicePortForwardResponse,
    responseDeserialize: deserialize_portal_daemon_api_RemoveUserServicePortForwardResponse,
  },
};

exports.KurtosisPortalDaemonClient = grpc.makeGenericClientConstructor(KurtosisPortalDaemonService);
