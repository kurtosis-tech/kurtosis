/**
 * @fileoverview gRPC-Web generated client stub for api_container_api
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
proto.api_container_api = require('./api_container_service_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.api_container_api.ApiContainerServiceClient =
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
proto.api_container_api.ApiContainerServicePromiseClient =
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
 *   !proto.api_container_api.LoadModuleArgs,
 *   !proto.api_container_api.LoadModuleResponse>}
 */
const methodDescriptor_ApiContainerService_LoadModule = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/LoadModule',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.LoadModuleArgs,
  proto.api_container_api.LoadModuleResponse,
  /**
   * @param {!proto.api_container_api.LoadModuleArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.LoadModuleResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.LoadModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.LoadModuleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.LoadModuleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.loadModule =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/LoadModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_LoadModule,
      callback);
};


/**
 * @param {!proto.api_container_api.LoadModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.LoadModuleResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.loadModule =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/LoadModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_LoadModule);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.UnloadModuleArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_UnloadModule = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/UnloadModule',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.UnloadModuleArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.UnloadModuleArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.UnloadModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.unloadModule =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UnloadModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UnloadModule,
      callback);
};


/**
 * @param {!proto.api_container_api.UnloadModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.unloadModule =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UnloadModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UnloadModule);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.ExecuteModuleArgs,
 *   !proto.api_container_api.ExecuteModuleResponse>}
 */
const methodDescriptor_ApiContainerService_ExecuteModule = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/ExecuteModule',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.ExecuteModuleArgs,
  proto.api_container_api.ExecuteModuleResponse,
  /**
   * @param {!proto.api_container_api.ExecuteModuleArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.ExecuteModuleResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.ExecuteModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.ExecuteModuleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.ExecuteModuleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.executeModule =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ExecuteModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ExecuteModule,
      callback);
};


/**
 * @param {!proto.api_container_api.ExecuteModuleArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.ExecuteModuleResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.executeModule =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ExecuteModule',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ExecuteModule);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.GetModuleInfoArgs,
 *   !proto.api_container_api.GetModuleInfoResponse>}
 */
const methodDescriptor_ApiContainerService_GetModuleInfo = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetModuleInfo',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.GetModuleInfoArgs,
  proto.api_container_api.GetModuleInfoResponse,
  /**
   * @param {!proto.api_container_api.GetModuleInfoArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetModuleInfoResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.GetModuleInfoArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.GetModuleInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.GetModuleInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.getModuleInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetModuleInfo',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetModuleInfo,
      callback);
};


/**
 * @param {!proto.api_container_api.GetModuleInfoArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.GetModuleInfoResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.getModuleInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetModuleInfo',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetModuleInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RegisterServiceArgs,
 *   !proto.api_container_api.RegisterServiceResponse>}
 */
const methodDescriptor_ApiContainerService_RegisterService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/RegisterService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.RegisterServiceArgs,
  proto.api_container_api.RegisterServiceResponse,
  /**
   * @param {!proto.api_container_api.RegisterServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.RegisterServiceResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RegisterServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.RegisterServiceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.RegisterServiceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.registerService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RegisterService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RegisterService,
      callback);
};


/**
 * @param {!proto.api_container_api.RegisterServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.RegisterServiceResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.registerService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RegisterService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RegisterService);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.StartServiceArgs,
 *   !proto.api_container_api.StartServiceResponse>}
 */
const methodDescriptor_ApiContainerService_StartService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/StartService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.StartServiceArgs,
  proto.api_container_api.StartServiceResponse,
  /**
   * @param {!proto.api_container_api.StartServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StartServiceResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.StartServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.StartServiceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StartServiceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.startService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StartService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StartService,
      callback);
};


/**
 * @param {!proto.api_container_api.StartServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.StartServiceResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.startService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StartService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StartService);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.GetServiceInfoArgs,
 *   !proto.api_container_api.GetServiceInfoResponse>}
 */
const methodDescriptor_ApiContainerService_GetServiceInfo = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetServiceInfo',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.GetServiceInfoArgs,
  proto.api_container_api.GetServiceInfoResponse,
  /**
   * @param {!proto.api_container_api.GetServiceInfoArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetServiceInfoResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.GetServiceInfoArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.GetServiceInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.GetServiceInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.getServiceInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetServiceInfo',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetServiceInfo,
      callback);
};


/**
 * @param {!proto.api_container_api.GetServiceInfoArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.GetServiceInfoResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.getServiceInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetServiceInfo',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetServiceInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RemoveServiceArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_RemoveService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/RemoveService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.RemoveServiceArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.RemoveServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RemoveServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.removeService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RemoveService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RemoveService,
      callback);
};


/**
 * @param {!proto.api_container_api.RemoveServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.removeService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RemoveService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RemoveService);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RepartitionArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_Repartition = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/Repartition',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.RepartitionArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.RepartitionArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RepartitionArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.repartition =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/Repartition',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_Repartition,
      callback);
};


/**
 * @param {!proto.api_container_api.RepartitionArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.repartition =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/Repartition',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_Repartition);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.ExecCommandArgs,
 *   !proto.api_container_api.ExecCommandResponse>}
 */
const methodDescriptor_ApiContainerService_ExecCommand = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/ExecCommand',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.ExecCommandArgs,
  proto.api_container_api.ExecCommandResponse,
  /**
   * @param {!proto.api_container_api.ExecCommandArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.ExecCommandResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.ExecCommandArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.ExecCommandResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.ExecCommandResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.execCommand =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ExecCommand',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ExecCommand,
      callback);
};


/**
 * @param {!proto.api_container_api.ExecCommandArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.ExecCommandResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.execCommand =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ExecCommand',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ExecCommand);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.PauseServiceArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_PauseService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/PauseService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.PauseServiceArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.PauseServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.PauseServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.pauseService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/PauseService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_PauseService,
      callback);
};


/**
 * @param {!proto.api_container_api.PauseServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.pauseService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/PauseService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_PauseService);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.UnpauseServiceArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_UnpauseService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/UnpauseService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.UnpauseServiceArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.UnpauseServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.UnpauseServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.unpauseService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UnpauseService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UnpauseService,
      callback);
};


/**
 * @param {!proto.api_container_api.UnpauseServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.unpauseService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UnpauseService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UnpauseService);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_WaitForHttpGetEndpointAvailability = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.waitForHttpGetEndpointAvailability =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_WaitForHttpGetEndpointAvailability,
      callback);
};


/**
 * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.waitForHttpGetEndpointAvailability =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_WaitForHttpGetEndpointAvailability);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_ApiContainerService_WaitForHttpPostEndpointAvailability = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.waitForHttpPostEndpointAvailability =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_WaitForHttpPostEndpointAvailability,
      callback);
};


/**
 * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.waitForHttpPostEndpointAvailability =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_WaitForHttpPostEndpointAvailability);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.api_container_api.GetServicesResponse>}
 */
const methodDescriptor_ApiContainerService_GetServices = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetServices',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.api_container_api.GetServicesResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetServicesResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.GetServicesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.GetServicesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.getServices =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetServices',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetServices,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.GetServicesResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.getServices =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetServices',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetServices);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.api_container_api.GetModulesResponse>}
 */
const methodDescriptor_ApiContainerService_GetModules = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetModules',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.api_container_api.GetModulesResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetModulesResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.GetModulesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.GetModulesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.getModules =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetModules',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetModules,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.GetModulesResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.getModules =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetModules',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetModules);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.UploadFilesArtifactArgs,
 *   !proto.api_container_api.UploadFilesArtifactResponse>}
 */
const methodDescriptor_ApiContainerService_UploadFilesArtifact = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/UploadFilesArtifact',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.UploadFilesArtifactArgs,
  proto.api_container_api.UploadFilesArtifactResponse,
  /**
   * @param {!proto.api_container_api.UploadFilesArtifactArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.UploadFilesArtifactResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.UploadFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.UploadFilesArtifactResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.UploadFilesArtifactResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.uploadFilesArtifact =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UploadFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UploadFilesArtifact,
      callback);
};


/**
 * @param {!proto.api_container_api.UploadFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.UploadFilesArtifactResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.uploadFilesArtifact =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/UploadFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_UploadFilesArtifact);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.StoreWebFilesArtifactArgs,
 *   !proto.api_container_api.StoreWebFilesArtifactResponse>}
 */
const methodDescriptor_ApiContainerService_StoreWebFilesArtifact = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/StoreWebFilesArtifact',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.StoreWebFilesArtifactArgs,
  proto.api_container_api.StoreWebFilesArtifactResponse,
  /**
   * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StoreWebFilesArtifactResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.StoreWebFilesArtifactResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StoreWebFilesArtifactResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.storeWebFilesArtifact =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StoreWebFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StoreWebFilesArtifact,
      callback);
};


/**
 * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.StoreWebFilesArtifactResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.storeWebFilesArtifact =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StoreWebFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StoreWebFilesArtifact);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.StoreFilesArtifactFromServiceArgs,
 *   !proto.api_container_api.StoreFilesArtifactFromServiceResponse>}
 */
const methodDescriptor_ApiContainerService_StoreFilesArtifactFromService = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/StoreFilesArtifactFromService',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.StoreFilesArtifactFromServiceArgs,
  proto.api_container_api.StoreFilesArtifactFromServiceResponse,
  /**
   * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StoreFilesArtifactFromServiceResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.StoreFilesArtifactFromServiceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StoreFilesArtifactFromServiceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.storeFilesArtifactFromService =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StoreFilesArtifactFromService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StoreFilesArtifactFromService,
      callback);
};


/**
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.StoreFilesArtifactFromServiceResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.storeFilesArtifactFromService =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/StoreFilesArtifactFromService',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_StoreFilesArtifactFromService);
};


module.exports = proto.api_container_api;

