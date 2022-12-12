/**
 * @fileoverview gRPC-Web generated client stub for module_api
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js')
const proto = {};
proto.module_api = require('./executable_module_service_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.module_api.ExecutableModuleServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.module_api.ExecutableModuleServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ExecutableModuleService_IsAvailable = new grpc.web.MethodDescriptor(
  '/module_api.ExecutableModuleService/IsAvailable',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.module_api.ExecutableModuleServiceClient.prototype.isAvailable =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/module_api.ExecutableModuleService/IsAvailable',
      request,
      metadata || {},
      methodDescriptor_ExecutableModuleService_IsAvailable,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.module_api.ExecutableModuleServicePromiseClient.prototype.isAvailable =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/module_api.ExecutableModuleService/IsAvailable',
      request,
      metadata || {},
      methodDescriptor_ExecutableModuleService_IsAvailable);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.module_api.ExecuteArgs,
 *   !proto.module_api.ExecuteResponse>}
 */
const methodDescriptor_ExecutableModuleService_Execute = new grpc.web.MethodDescriptor(
  '/module_api.ExecutableModuleService/Execute',
  grpc.web.MethodType.UNARY,
  proto.module_api.ExecuteArgs,
  proto.module_api.ExecuteResponse,
  /**
   * @param {!proto.module_api.ExecuteArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.module_api.ExecuteResponse.deserializeBinary
);


/**
 * @param {!proto.module_api.ExecuteArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.module_api.ExecuteResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.module_api.ExecuteResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.module_api.ExecutableModuleServiceClient.prototype.execute =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/module_api.ExecutableModuleService/Execute',
      request,
      metadata || {},
      methodDescriptor_ExecutableModuleService_Execute,
      callback);
};


/**
 * @param {!proto.module_api.ExecuteArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.module_api.ExecuteResponse>}
 *     Promise that resolves to the response
 */
proto.module_api.ExecutableModuleServicePromiseClient.prototype.execute =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/module_api.ExecutableModuleService/Execute',
      request,
      metadata || {},
      methodDescriptor_ExecutableModuleService_Execute);
};


module.exports = proto.module_api;

