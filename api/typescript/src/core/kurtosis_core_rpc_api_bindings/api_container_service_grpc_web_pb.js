/**
 * @fileoverview gRPC-Web generated client stub for api_container_api
 * @enhanceable
 * @public
 */

// Code generated by protoc-gen-grpc-web. DO NOT EDIT.
// versions:
// 	protoc-gen-grpc-web v1.4.2
// 	protoc              v3.19.1
// source: api_container_service.proto


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
  this.hostname_ = hostname.replace(/\/+$/, '');

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
  this.hostname_ = hostname.replace(/\/+$/, '');

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RunStarlarkScriptArgs,
 *   !proto.api_container_api.StarlarkRunResponseLine>}
 */
const methodDescriptor_ApiContainerService_RunStarlarkScript = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/RunStarlarkScript',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.api_container_api.RunStarlarkScriptArgs,
  proto.api_container_api.StarlarkRunResponseLine,
  /**
   * @param {!proto.api_container_api.RunStarlarkScriptArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StarlarkRunResponseLine.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RunStarlarkScriptArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StarlarkRunResponseLine>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.runStarlarkScript =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/RunStarlarkScript',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RunStarlarkScript);
};


/**
 * @param {!proto.api_container_api.RunStarlarkScriptArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StarlarkRunResponseLine>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.runStarlarkScript =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/RunStarlarkScript',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RunStarlarkScript);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RunStarlarkPackageArgs,
 *   !proto.api_container_api.StarlarkRunResponseLine>}
 */
const methodDescriptor_ApiContainerService_RunStarlarkPackage = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/RunStarlarkPackage',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.api_container_api.RunStarlarkPackageArgs,
  proto.api_container_api.StarlarkRunResponseLine,
  /**
   * @param {!proto.api_container_api.RunStarlarkPackageArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StarlarkRunResponseLine.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RunStarlarkPackageArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StarlarkRunResponseLine>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.runStarlarkPackage =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/RunStarlarkPackage',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RunStarlarkPackage);
};


/**
 * @param {!proto.api_container_api.RunStarlarkPackageArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StarlarkRunResponseLine>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.runStarlarkPackage =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/RunStarlarkPackage',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RunStarlarkPackage);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.GetServicesArgs,
 *   !proto.api_container_api.GetServicesResponse>}
 */
const methodDescriptor_ApiContainerService_GetServices = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetServices',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.GetServicesArgs,
  proto.api_container_api.GetServicesResponse,
  /**
   * @param {!proto.api_container_api.GetServicesArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetServicesResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.GetServicesArgs} request The
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
 * @param {!proto.api_container_api.GetServicesArgs} request The
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
 *   !proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse>}
 */
const methodDescriptor_ApiContainerService_GetExistingAndHistoricalServiceIdentifiers = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.getExistingAndHistoricalServiceIdentifiers =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetExistingAndHistoricalServiceIdentifiers,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.getExistingAndHistoricalServiceIdentifiers =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_GetExistingAndHistoricalServiceIdentifiers);
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
 *   !proto.api_container_api.DownloadFilesArtifactArgs,
 *   !proto.api_container_api.DownloadFilesArtifactResponse>}
 */
const methodDescriptor_ApiContainerService_DownloadFilesArtifact = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/DownloadFilesArtifact',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.DownloadFilesArtifactArgs,
  proto.api_container_api.DownloadFilesArtifactResponse,
  /**
   * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.DownloadFilesArtifactResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.DownloadFilesArtifactResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.DownloadFilesArtifactResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.downloadFilesArtifact =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/DownloadFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_DownloadFilesArtifact,
      callback);
};


/**
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.DownloadFilesArtifactResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.downloadFilesArtifact =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/DownloadFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_DownloadFilesArtifact);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.DownloadFilesArtifactArgs,
 *   !proto.api_container_api.StreamedDataChunk>}
 */
const methodDescriptor_ApiContainerService_DownloadFilesArtifactV2 = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/DownloadFilesArtifactV2',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.api_container_api.DownloadFilesArtifactArgs,
  proto.api_container_api.StreamedDataChunk,
  /**
   * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.StreamedDataChunk.deserializeBinary
);


/**
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StreamedDataChunk>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.downloadFilesArtifactV2 =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/DownloadFilesArtifactV2',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_DownloadFilesArtifactV2);
};


/**
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.StreamedDataChunk>}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.downloadFilesArtifactV2 =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/api_container_api.ApiContainerService/DownloadFilesArtifactV2',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_DownloadFilesArtifactV2);
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


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.api_container_api.RenderTemplatesToFilesArtifactArgs,
 *   !proto.api_container_api.RenderTemplatesToFilesArtifactResponse>}
 */
const methodDescriptor_ApiContainerService_RenderTemplatesToFilesArtifact = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact',
  grpc.web.MethodType.UNARY,
  proto.api_container_api.RenderTemplatesToFilesArtifactArgs,
  proto.api_container_api.RenderTemplatesToFilesArtifactResponse,
  /**
   * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.RenderTemplatesToFilesArtifactResponse.deserializeBinary
);


/**
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.RenderTemplatesToFilesArtifactResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.RenderTemplatesToFilesArtifactResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.renderTemplatesToFilesArtifact =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RenderTemplatesToFilesArtifact,
      callback);
};


/**
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.RenderTemplatesToFilesArtifactResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.renderTemplatesToFilesArtifact =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_RenderTemplatesToFilesArtifact);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse>}
 */
const methodDescriptor_ApiContainerService_ListFilesArtifactNamesAndUuids = new grpc.web.MethodDescriptor(
  '/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.api_container_api.ApiContainerServiceClient.prototype.listFilesArtifactNamesAndUuids =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ListFilesArtifactNamesAndUuids,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse>}
 *     Promise that resolves to the response
 */
proto.api_container_api.ApiContainerServicePromiseClient.prototype.listFilesArtifactNamesAndUuids =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids',
      request,
      metadata || {},
      methodDescriptor_ApiContainerService_ListFilesArtifactNamesAndUuids);
};


module.exports = proto.api_container_api;

