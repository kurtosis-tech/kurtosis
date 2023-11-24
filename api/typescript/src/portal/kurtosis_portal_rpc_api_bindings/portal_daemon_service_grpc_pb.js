// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var portal_daemon_service_pb = require('./portal_daemon_service_pb.js');

function serialize_portal_daemon_api_ForwardUserServicePortArgs(arg) {
  if (!(arg instanceof portal_daemon_service_pb.ForwardUserServicePortArgs)) {
    throw new Error('Expected argument of type portal_daemon_api.ForwardUserServicePortArgs');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_ForwardUserServicePortArgs(buffer_arg) {
  return portal_daemon_service_pb.ForwardUserServicePortArgs.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_portal_daemon_api_ForwardUserServicePortResponse(arg) {
  if (!(arg instanceof portal_daemon_service_pb.ForwardUserServicePortResponse)) {
    throw new Error('Expected argument of type portal_daemon_api.ForwardUserServicePortResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_portal_daemon_api_ForwardUserServicePortResponse(buffer_arg) {
  return portal_daemon_service_pb.ForwardUserServicePortResponse.deserializeBinary(new Uint8Array(buffer_arg));
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
  forwardUserServicePort: {
    path: '/portal_daemon_api.KurtosisPortalDaemon/ForwardUserServicePort',
    requestStream: false,
    responseStream: false,
    requestType: portal_daemon_service_pb.ForwardUserServicePortArgs,
    responseType: portal_daemon_service_pb.ForwardUserServicePortResponse,
    requestSerialize: serialize_portal_daemon_api_ForwardUserServicePortArgs,
    requestDeserialize: deserialize_portal_daemon_api_ForwardUserServicePortArgs,
    responseSerialize: serialize_portal_daemon_api_ForwardUserServicePortResponse,
    responseDeserialize: deserialize_portal_daemon_api_ForwardUserServicePortResponse,
  },
};

exports.KurtosisPortalDaemonClient = grpc.makeGenericClientConstructor(KurtosisPortalDaemonService);
