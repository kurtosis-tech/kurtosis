// source: api_container_service.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
goog.object.extend(proto, google_protobuf_empty_pb);
goog.exportSymbol('proto.api_container_api.DataChunkMetadata', null, global);
goog.exportSymbol('proto.api_container_api.DownloadFilesArtifactArgs', null, global);
goog.exportSymbol('proto.api_container_api.DownloadFilesArtifactResponse', null, global);
goog.exportSymbol('proto.api_container_api.ExecCommandArgs', null, global);
goog.exportSymbol('proto.api_container_api.ExecCommandResponse', null, global);
goog.exportSymbol('proto.api_container_api.FilesArtifactNameAndUuid', null, global);
goog.exportSymbol('proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse', null, global);
goog.exportSymbol('proto.api_container_api.GetServicesArgs', null, global);
goog.exportSymbol('proto.api_container_api.GetServicesResponse', null, global);
goog.exportSymbol('proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse', null, global);
goog.exportSymbol('proto.api_container_api.PartitionConnectionInfo', null, global);
goog.exportSymbol('proto.api_container_api.PartitionConnections', null, global);
goog.exportSymbol('proto.api_container_api.PartitionServices', null, global);
goog.exportSymbol('proto.api_container_api.PauseServiceArgs', null, global);
goog.exportSymbol('proto.api_container_api.Port', null, global);
goog.exportSymbol('proto.api_container_api.Port.TransportProtocol', null, global);
goog.exportSymbol('proto.api_container_api.RemoveServiceArgs', null, global);
goog.exportSymbol('proto.api_container_api.RemoveServiceResponse', null, global);
goog.exportSymbol('proto.api_container_api.RenderTemplatesToFilesArtifactArgs', null, global);
goog.exportSymbol('proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData', null, global);
goog.exportSymbol('proto.api_container_api.RenderTemplatesToFilesArtifactResponse', null, global);
goog.exportSymbol('proto.api_container_api.RepartitionArgs', null, global);
goog.exportSymbol('proto.api_container_api.RunStarlarkPackageArgs', null, global);
goog.exportSymbol('proto.api_container_api.RunStarlarkPackageArgs.StarlarkPackageContentCase', null, global);
goog.exportSymbol('proto.api_container_api.RunStarlarkScriptArgs', null, global);
goog.exportSymbol('proto.api_container_api.ServiceConfig', null, global);
goog.exportSymbol('proto.api_container_api.ServiceIdentifiers', null, global);
goog.exportSymbol('proto.api_container_api.ServiceInfo', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkError', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkError.ErrorCase', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkExecutionError', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkInstruction', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkInstructionArg', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkInstructionPosition', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkInstructionResult', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkInterpretationError', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkRunFinishedEvent', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkRunProgress', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkRunResponseLine', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkRunResponseLine.RunResponseLineCase', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkValidationError', null, global);
goog.exportSymbol('proto.api_container_api.StarlarkWarning', null, global);
goog.exportSymbol('proto.api_container_api.StartServicesArgs', null, global);
goog.exportSymbol('proto.api_container_api.StartServicesResponse', null, global);
goog.exportSymbol('proto.api_container_api.StoreFilesArtifactFromServiceArgs', null, global);
goog.exportSymbol('proto.api_container_api.StoreFilesArtifactFromServiceResponse', null, global);
goog.exportSymbol('proto.api_container_api.StoreWebFilesArtifactArgs', null, global);
goog.exportSymbol('proto.api_container_api.StoreWebFilesArtifactResponse', null, global);
goog.exportSymbol('proto.api_container_api.StreamedDataChunk', null, global);
goog.exportSymbol('proto.api_container_api.UnpauseServiceArgs', null, global);
goog.exportSymbol('proto.api_container_api.UpdateServiceConfig', null, global);
goog.exportSymbol('proto.api_container_api.UploadFilesArtifactArgs', null, global);
goog.exportSymbol('proto.api_container_api.UploadFilesArtifactResponse', null, global);
goog.exportSymbol('proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs', null, global);
goog.exportSymbol('proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.Port = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.Port, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.Port.displayName = 'proto.api_container_api.Port';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ServiceInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.ServiceInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ServiceInfo.displayName = 'proto.api_container_api.ServiceInfo';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ServiceConfig = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.ServiceConfig.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.ServiceConfig, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ServiceConfig.displayName = 'proto.api_container_api.ServiceConfig';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.UpdateServiceConfig = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.UpdateServiceConfig, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.UpdateServiceConfig.displayName = 'proto.api_container_api.UpdateServiceConfig';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RunStarlarkScriptArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RunStarlarkScriptArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RunStarlarkScriptArgs.displayName = 'proto.api_container_api.RunStarlarkScriptArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RunStarlarkPackageArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_);
};
goog.inherits(proto.api_container_api.RunStarlarkPackageArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RunStarlarkPackageArgs.displayName = 'proto.api_container_api.RunStarlarkPackageArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkRunResponseLine = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_);
};
goog.inherits(proto.api_container_api.StarlarkRunResponseLine, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkRunResponseLine.displayName = 'proto.api_container_api.StarlarkRunResponseLine';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkWarning = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkWarning, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkWarning.displayName = 'proto.api_container_api.StarlarkWarning';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkInstruction = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.StarlarkInstruction.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.StarlarkInstruction, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkInstruction.displayName = 'proto.api_container_api.StarlarkInstruction';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkInstructionResult = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkInstructionResult, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkInstructionResult.displayName = 'proto.api_container_api.StarlarkInstructionResult';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkInstructionArg = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkInstructionArg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkInstructionArg.displayName = 'proto.api_container_api.StarlarkInstructionArg';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkInstructionPosition = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkInstructionPosition, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkInstructionPosition.displayName = 'proto.api_container_api.StarlarkInstructionPosition';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkError = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.api_container_api.StarlarkError.oneofGroups_);
};
goog.inherits(proto.api_container_api.StarlarkError, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkError.displayName = 'proto.api_container_api.StarlarkError';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkInterpretationError = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkInterpretationError, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkInterpretationError.displayName = 'proto.api_container_api.StarlarkInterpretationError';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkValidationError = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkValidationError, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkValidationError.displayName = 'proto.api_container_api.StarlarkValidationError';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkExecutionError = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkExecutionError, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkExecutionError.displayName = 'proto.api_container_api.StarlarkExecutionError';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkRunProgress = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.StarlarkRunProgress.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.StarlarkRunProgress, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkRunProgress.displayName = 'proto.api_container_api.StarlarkRunProgress';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StarlarkRunFinishedEvent = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StarlarkRunFinishedEvent, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StarlarkRunFinishedEvent.displayName = 'proto.api_container_api.StarlarkRunFinishedEvent';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StartServicesArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StartServicesArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StartServicesArgs.displayName = 'proto.api_container_api.StartServicesArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StartServicesResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StartServicesResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StartServicesResponse.displayName = 'proto.api_container_api.StartServicesResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.GetServicesArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.GetServicesArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.GetServicesArgs.displayName = 'proto.api_container_api.GetServicesArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.GetServicesResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.GetServicesResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.GetServicesResponse.displayName = 'proto.api_container_api.GetServicesResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ServiceIdentifiers = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.ServiceIdentifiers, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ServiceIdentifiers.displayName = 'proto.api_container_api.ServiceIdentifiers';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.displayName = 'proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RemoveServiceArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RemoveServiceArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RemoveServiceArgs.displayName = 'proto.api_container_api.RemoveServiceArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RemoveServiceResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RemoveServiceResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RemoveServiceResponse.displayName = 'proto.api_container_api.RemoveServiceResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RepartitionArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RepartitionArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RepartitionArgs.displayName = 'proto.api_container_api.RepartitionArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.PartitionServices = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.PartitionServices, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.PartitionServices.displayName = 'proto.api_container_api.PartitionServices';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.PartitionConnections = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.PartitionConnections, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.PartitionConnections.displayName = 'proto.api_container_api.PartitionConnections';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.PartitionConnectionInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.PartitionConnectionInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.PartitionConnectionInfo.displayName = 'proto.api_container_api.PartitionConnectionInfo';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ExecCommandArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.ExecCommandArgs.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.ExecCommandArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ExecCommandArgs.displayName = 'proto.api_container_api.ExecCommandArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.PauseServiceArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.PauseServiceArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.PauseServiceArgs.displayName = 'proto.api_container_api.PauseServiceArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.UnpauseServiceArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.UnpauseServiceArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.UnpauseServiceArgs.displayName = 'proto.api_container_api.UnpauseServiceArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ExecCommandResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.ExecCommandResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ExecCommandResponse.displayName = 'proto.api_container_api.ExecCommandResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.displayName = 'proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.displayName = 'proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StreamedDataChunk = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StreamedDataChunk, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StreamedDataChunk.displayName = 'proto.api_container_api.StreamedDataChunk';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.DataChunkMetadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.DataChunkMetadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.DataChunkMetadata.displayName = 'proto.api_container_api.DataChunkMetadata';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.UploadFilesArtifactArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.UploadFilesArtifactArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.UploadFilesArtifactArgs.displayName = 'proto.api_container_api.UploadFilesArtifactArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.UploadFilesArtifactResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.UploadFilesArtifactResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.UploadFilesArtifactResponse.displayName = 'proto.api_container_api.UploadFilesArtifactResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.DownloadFilesArtifactArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.DownloadFilesArtifactArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.DownloadFilesArtifactArgs.displayName = 'proto.api_container_api.DownloadFilesArtifactArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.DownloadFilesArtifactResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.DownloadFilesArtifactResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.DownloadFilesArtifactResponse.displayName = 'proto.api_container_api.DownloadFilesArtifactResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StoreWebFilesArtifactArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StoreWebFilesArtifactArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StoreWebFilesArtifactArgs.displayName = 'proto.api_container_api.StoreWebFilesArtifactArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StoreWebFilesArtifactResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StoreWebFilesArtifactResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StoreWebFilesArtifactResponse.displayName = 'proto.api_container_api.StoreWebFilesArtifactResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StoreFilesArtifactFromServiceArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StoreFilesArtifactFromServiceArgs.displayName = 'proto.api_container_api.StoreFilesArtifactFromServiceArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.StoreFilesArtifactFromServiceResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.StoreFilesArtifactFromServiceResponse.displayName = 'proto.api_container_api.StoreFilesArtifactFromServiceResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RenderTemplatesToFilesArtifactArgs, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RenderTemplatesToFilesArtifactArgs.displayName = 'proto.api_container_api.RenderTemplatesToFilesArtifactArgs';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.displayName = 'proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.RenderTemplatesToFilesArtifactResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.RenderTemplatesToFilesArtifactResponse.displayName = 'proto.api_container_api.RenderTemplatesToFilesArtifactResponse';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.FilesArtifactNameAndUuid = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.api_container_api.FilesArtifactNameAndUuid, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.FilesArtifactNameAndUuid.displayName = 'proto.api_container_api.FilesArtifactNameAndUuid';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.repeatedFields_, null);
};
goog.inherits(proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.displayName = 'proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.Port.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.Port.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.Port} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.Port.toObject = function(includeInstance, msg) {
  var f, obj = {
    number: jspb.Message.getFieldWithDefault(msg, 1, 0),
    transportProtocol: jspb.Message.getFieldWithDefault(msg, 2, 0),
    maybeApplicationProtocol: jspb.Message.getFieldWithDefault(msg, 3, ""),
    maybeWaitTimeout: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.Port}
 */
proto.api_container_api.Port.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.Port;
  return proto.api_container_api.Port.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.Port} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.Port}
 */
proto.api_container_api.Port.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setNumber(value);
      break;
    case 2:
      var value = /** @type {!proto.api_container_api.Port.TransportProtocol} */ (reader.readEnum());
      msg.setTransportProtocol(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMaybeApplicationProtocol(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setMaybeWaitTimeout(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.Port.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.Port.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.Port} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.Port.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNumber();
  if (f !== 0) {
    writer.writeUint32(
      1,
      f
    );
  }
  f = message.getTransportProtocol();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getMaybeApplicationProtocol();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getMaybeWaitTimeout();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * @enum {number}
 */
proto.api_container_api.Port.TransportProtocol = {
  TCP: 0,
  SCTP: 1,
  UDP: 2
};

/**
 * optional uint32 number = 1;
 * @return {number}
 */
proto.api_container_api.Port.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.Port} returns this
 */
proto.api_container_api.Port.prototype.setNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional TransportProtocol transport_protocol = 2;
 * @return {!proto.api_container_api.Port.TransportProtocol}
 */
proto.api_container_api.Port.prototype.getTransportProtocol = function() {
  return /** @type {!proto.api_container_api.Port.TransportProtocol} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.api_container_api.Port.TransportProtocol} value
 * @return {!proto.api_container_api.Port} returns this
 */
proto.api_container_api.Port.prototype.setTransportProtocol = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string maybe_application_protocol = 3;
 * @return {string}
 */
proto.api_container_api.Port.prototype.getMaybeApplicationProtocol = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.Port} returns this
 */
proto.api_container_api.Port.prototype.setMaybeApplicationProtocol = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string maybe_wait_timeout = 4;
 * @return {string}
 */
proto.api_container_api.Port.prototype.getMaybeWaitTimeout = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.Port} returns this
 */
proto.api_container_api.Port.prototype.setMaybeWaitTimeout = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ServiceInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ServiceInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ServiceInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceUuid: jspb.Message.getFieldWithDefault(msg, 1, ""),
    privateIpAddr: jspb.Message.getFieldWithDefault(msg, 2, ""),
    privatePortsMap: (f = msg.getPrivatePortsMap()) ? f.toObject(includeInstance, proto.api_container_api.Port.toObject) : [],
    maybePublicIpAddr: jspb.Message.getFieldWithDefault(msg, 4, ""),
    maybePublicPortsMap: (f = msg.getMaybePublicPortsMap()) ? f.toObject(includeInstance, proto.api_container_api.Port.toObject) : [],
    name: jspb.Message.getFieldWithDefault(msg, 6, ""),
    shortenedUuid: jspb.Message.getFieldWithDefault(msg, 7, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ServiceInfo}
 */
proto.api_container_api.ServiceInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ServiceInfo;
  return proto.api_container_api.ServiceInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ServiceInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ServiceInfo}
 */
proto.api_container_api.ServiceInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceUuid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPrivateIpAddr(value);
      break;
    case 3:
      var value = msg.getPrivatePortsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.Port.deserializeBinaryFromReader, "", new proto.api_container_api.Port());
         });
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setMaybePublicIpAddr(value);
      break;
    case 5:
      var value = msg.getMaybePublicPortsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.Port.deserializeBinaryFromReader, "", new proto.api_container_api.Port());
         });
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setShortenedUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ServiceInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ServiceInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ServiceInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPrivateIpAddr();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getPrivatePortsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(3, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.Port.serializeBinaryToWriter);
  }
  f = message.getMaybePublicIpAddr();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getMaybePublicPortsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(5, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.Port.serializeBinaryToWriter);
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getShortenedUuid();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
};


/**
 * optional string service_uuid = 1;
 * @return {string}
 */
proto.api_container_api.ServiceInfo.prototype.getServiceUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.setServiceUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string private_ip_addr = 2;
 * @return {string}
 */
proto.api_container_api.ServiceInfo.prototype.getPrivateIpAddr = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.setPrivateIpAddr = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * map<string, Port> private_ports = 3;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.Port>}
 */
proto.api_container_api.ServiceInfo.prototype.getPrivatePortsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.Port>} */ (
      jspb.Message.getMapField(this, 3, opt_noLazyCreate,
      proto.api_container_api.Port));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.clearPrivatePortsMap = function() {
  this.getPrivatePortsMap().clear();
  return this;};


/**
 * optional string maybe_public_ip_addr = 4;
 * @return {string}
 */
proto.api_container_api.ServiceInfo.prototype.getMaybePublicIpAddr = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.setMaybePublicIpAddr = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * map<string, Port> maybe_public_ports = 5;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.Port>}
 */
proto.api_container_api.ServiceInfo.prototype.getMaybePublicPortsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.Port>} */ (
      jspb.Message.getMapField(this, 5, opt_noLazyCreate,
      proto.api_container_api.Port));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.clearMaybePublicPortsMap = function() {
  this.getMaybePublicPortsMap().clear();
  return this;};


/**
 * optional string name = 6;
 * @return {string}
 */
proto.api_container_api.ServiceInfo.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string shortened_uuid = 7;
 * @return {string}
 */
proto.api_container_api.ServiceInfo.prototype.getShortenedUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceInfo} returns this
 */
proto.api_container_api.ServiceInfo.prototype.setShortenedUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.ServiceConfig.repeatedFields_ = [4,5];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ServiceConfig.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ServiceConfig.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ServiceConfig} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceConfig.toObject = function(includeInstance, msg) {
  var f, obj = {
    containerImageName: jspb.Message.getFieldWithDefault(msg, 1, ""),
    privatePortsMap: (f = msg.getPrivatePortsMap()) ? f.toObject(includeInstance, proto.api_container_api.Port.toObject) : [],
    publicPortsMap: (f = msg.getPublicPortsMap()) ? f.toObject(includeInstance, proto.api_container_api.Port.toObject) : [],
    entrypointArgsList: (f = jspb.Message.getRepeatedField(msg, 4)) == null ? undefined : f,
    cmdArgsList: (f = jspb.Message.getRepeatedField(msg, 5)) == null ? undefined : f,
    envVarsMap: (f = msg.getEnvVarsMap()) ? f.toObject(includeInstance, undefined) : [],
    filesArtifactMountpointsMap: (f = msg.getFilesArtifactMountpointsMap()) ? f.toObject(includeInstance, undefined) : [],
    cpuAllocationMillicpus: jspb.Message.getFieldWithDefault(msg, 8, 0),
    memoryAllocationMegabytes: jspb.Message.getFieldWithDefault(msg, 9, 0),
    privateIpAddrPlaceholder: jspb.Message.getFieldWithDefault(msg, 10, ""),
    subnetwork: jspb.Message.getFieldWithDefault(msg, 11, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ServiceConfig}
 */
proto.api_container_api.ServiceConfig.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ServiceConfig;
  return proto.api_container_api.ServiceConfig.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ServiceConfig} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ServiceConfig}
 */
proto.api_container_api.ServiceConfig.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setContainerImageName(value);
      break;
    case 2:
      var value = msg.getPrivatePortsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.Port.deserializeBinaryFromReader, "", new proto.api_container_api.Port());
         });
      break;
    case 3:
      var value = msg.getPublicPortsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.Port.deserializeBinaryFromReader, "", new proto.api_container_api.Port());
         });
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.addEntrypointArgs(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.addCmdArgs(value);
      break;
    case 6:
      var value = msg.getEnvVarsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 7:
      var value = msg.getFilesArtifactMountpointsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 8:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setCpuAllocationMillicpus(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setMemoryAllocationMegabytes(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setPrivateIpAddrPlaceholder(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setSubnetwork(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ServiceConfig.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ServiceConfig.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ServiceConfig} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceConfig.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getContainerImageName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPrivatePortsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.Port.serializeBinaryToWriter);
  }
  f = message.getPublicPortsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(3, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.Port.serializeBinaryToWriter);
  }
  f = message.getEntrypointArgsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      4,
      f
    );
  }
  f = message.getCmdArgsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      5,
      f
    );
  }
  f = message.getEnvVarsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(6, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getFilesArtifactMountpointsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(7, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getCpuAllocationMillicpus();
  if (f !== 0) {
    writer.writeUint64(
      8,
      f
    );
  }
  f = message.getMemoryAllocationMegabytes();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getPrivateIpAddrPlaceholder();
  if (f.length > 0) {
    writer.writeString(
      10,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 11));
  if (f != null) {
    writer.writeString(
      11,
      f
    );
  }
};


/**
 * optional string container_image_name = 1;
 * @return {string}
 */
proto.api_container_api.ServiceConfig.prototype.getContainerImageName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setContainerImageName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * map<string, Port> private_ports = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.Port>}
 */
proto.api_container_api.ServiceConfig.prototype.getPrivatePortsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.Port>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      proto.api_container_api.Port));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearPrivatePortsMap = function() {
  this.getPrivatePortsMap().clear();
  return this;};


/**
 * map<string, Port> public_ports = 3;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.Port>}
 */
proto.api_container_api.ServiceConfig.prototype.getPublicPortsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.Port>} */ (
      jspb.Message.getMapField(this, 3, opt_noLazyCreate,
      proto.api_container_api.Port));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearPublicPortsMap = function() {
  this.getPublicPortsMap().clear();
  return this;};


/**
 * repeated string entrypoint_args = 4;
 * @return {!Array<string>}
 */
proto.api_container_api.ServiceConfig.prototype.getEntrypointArgsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 4));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setEntrypointArgsList = function(value) {
  return jspb.Message.setField(this, 4, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.addEntrypointArgs = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 4, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearEntrypointArgsList = function() {
  return this.setEntrypointArgsList([]);
};


/**
 * repeated string cmd_args = 5;
 * @return {!Array<string>}
 */
proto.api_container_api.ServiceConfig.prototype.getCmdArgsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 5));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setCmdArgsList = function(value) {
  return jspb.Message.setField(this, 5, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.addCmdArgs = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 5, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearCmdArgsList = function() {
  return this.setCmdArgsList([]);
};


/**
 * map<string, string> env_vars = 6;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api_container_api.ServiceConfig.prototype.getEnvVarsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 6, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearEnvVarsMap = function() {
  this.getEnvVarsMap().clear();
  return this;};


/**
 * map<string, string> files_artifact_mountpoints = 7;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api_container_api.ServiceConfig.prototype.getFilesArtifactMountpointsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 7, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearFilesArtifactMountpointsMap = function() {
  this.getFilesArtifactMountpointsMap().clear();
  return this;};


/**
 * optional uint64 cpu_allocation_millicpus = 8;
 * @return {number}
 */
proto.api_container_api.ServiceConfig.prototype.getCpuAllocationMillicpus = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setCpuAllocationMillicpus = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional uint64 memory_allocation_megabytes = 9;
 * @return {number}
 */
proto.api_container_api.ServiceConfig.prototype.getMemoryAllocationMegabytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setMemoryAllocationMegabytes = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional string private_ip_addr_placeholder = 10;
 * @return {string}
 */
proto.api_container_api.ServiceConfig.prototype.getPrivateIpAddrPlaceholder = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setPrivateIpAddrPlaceholder = function(value) {
  return jspb.Message.setProto3StringField(this, 10, value);
};


/**
 * optional string subnetwork = 11;
 * @return {string}
 */
proto.api_container_api.ServiceConfig.prototype.getSubnetwork = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.setSubnetwork = function(value) {
  return jspb.Message.setField(this, 11, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.ServiceConfig} returns this
 */
proto.api_container_api.ServiceConfig.prototype.clearSubnetwork = function() {
  return jspb.Message.setField(this, 11, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.ServiceConfig.prototype.hasSubnetwork = function() {
  return jspb.Message.getField(this, 11) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.UpdateServiceConfig.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.UpdateServiceConfig.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.UpdateServiceConfig} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UpdateServiceConfig.toObject = function(includeInstance, msg) {
  var f, obj = {
    subnetwork: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.UpdateServiceConfig}
 */
proto.api_container_api.UpdateServiceConfig.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.UpdateServiceConfig;
  return proto.api_container_api.UpdateServiceConfig.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.UpdateServiceConfig} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.UpdateServiceConfig}
 */
proto.api_container_api.UpdateServiceConfig.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSubnetwork(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.UpdateServiceConfig.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.UpdateServiceConfig.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.UpdateServiceConfig} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UpdateServiceConfig.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = /** @type {string} */ (jspb.Message.getField(message, 1));
  if (f != null) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string subnetwork = 1;
 * @return {string}
 */
proto.api_container_api.UpdateServiceConfig.prototype.getSubnetwork = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.UpdateServiceConfig} returns this
 */
proto.api_container_api.UpdateServiceConfig.prototype.setSubnetwork = function(value) {
  return jspb.Message.setField(this, 1, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.UpdateServiceConfig} returns this
 */
proto.api_container_api.UpdateServiceConfig.prototype.clearSubnetwork = function() {
  return jspb.Message.setField(this, 1, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.UpdateServiceConfig.prototype.hasSubnetwork = function() {
  return jspb.Message.getField(this, 1) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RunStarlarkScriptArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RunStarlarkScriptArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RunStarlarkScriptArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serializedScript: jspb.Message.getFieldWithDefault(msg, 1, ""),
    serializedParams: jspb.Message.getFieldWithDefault(msg, 2, ""),
    dryRun: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
    parallelism: jspb.Message.getFieldWithDefault(msg, 4, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RunStarlarkScriptArgs}
 */
proto.api_container_api.RunStarlarkScriptArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RunStarlarkScriptArgs;
  return proto.api_container_api.RunStarlarkScriptArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RunStarlarkScriptArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RunStarlarkScriptArgs}
 */
proto.api_container_api.RunStarlarkScriptArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedScript(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedParams(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDryRun(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setParallelism(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RunStarlarkScriptArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RunStarlarkScriptArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RunStarlarkScriptArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSerializedScript();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSerializedParams();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 3));
  if (f != null) {
    writer.writeBool(
      3,
      f
    );
  }
  f = /** @type {number} */ (jspb.Message.getField(message, 4));
  if (f != null) {
    writer.writeInt32(
      4,
      f
    );
  }
};


/**
 * optional string serialized_script = 1;
 * @return {string}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.getSerializedScript = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.setSerializedScript = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string serialized_params = 2;
 * @return {string}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.getSerializedParams = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.setSerializedParams = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bool dry_run = 3;
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.getDryRun = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.setDryRun = function(value) {
  return jspb.Message.setField(this, 3, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.clearDryRun = function() {
  return jspb.Message.setField(this, 3, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.hasDryRun = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional int32 parallelism = 4;
 * @return {number}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.getParallelism = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.setParallelism = function(value) {
  return jspb.Message.setField(this, 4, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkScriptArgs} returns this
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.clearParallelism = function() {
  return jspb.Message.setField(this, 4, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkScriptArgs.prototype.hasParallelism = function() {
  return jspb.Message.getField(this, 4) != null;
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_ = [[3,4]];

/**
 * @enum {number}
 */
proto.api_container_api.RunStarlarkPackageArgs.StarlarkPackageContentCase = {
  STARLARK_PACKAGE_CONTENT_NOT_SET: 0,
  LOCAL: 3,
  REMOTE: 4
};

/**
 * @return {proto.api_container_api.RunStarlarkPackageArgs.StarlarkPackageContentCase}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getStarlarkPackageContentCase = function() {
  return /** @type {proto.api_container_api.RunStarlarkPackageArgs.StarlarkPackageContentCase} */(jspb.Message.computeOneofCase(this, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_[0]));
};



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RunStarlarkPackageArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RunStarlarkPackageArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RunStarlarkPackageArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    packageId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    local: msg.getLocal_asB64(),
    remote: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    serializedParams: jspb.Message.getFieldWithDefault(msg, 5, ""),
    dryRun: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    parallelism: jspb.Message.getFieldWithDefault(msg, 7, 0),
    clonePackage: jspb.Message.getBooleanFieldWithDefault(msg, 8, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs}
 */
proto.api_container_api.RunStarlarkPackageArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RunStarlarkPackageArgs;
  return proto.api_container_api.RunStarlarkPackageArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RunStarlarkPackageArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs}
 */
proto.api_container_api.RunStarlarkPackageArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPackageId(value);
      break;
    case 3:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setLocal(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setRemote(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedParams(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDryRun(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setParallelism(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setClonePackage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RunStarlarkPackageArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RunStarlarkPackageArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RunStarlarkPackageArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPackageId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = /** @type {!(string|Uint8Array)} */ (jspb.Message.getField(message, 3));
  if (f != null) {
    writer.writeBytes(
      3,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 4));
  if (f != null) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getSerializedParams();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 6));
  if (f != null) {
    writer.writeBool(
      6,
      f
    );
  }
  f = /** @type {number} */ (jspb.Message.getField(message, 7));
  if (f != null) {
    writer.writeInt32(
      7,
      f
    );
  }
  f = /** @type {boolean} */ (jspb.Message.getField(message, 8));
  if (f != null) {
    writer.writeBool(
      8,
      f
    );
  }
};


/**
 * optional string package_id = 1;
 * @return {string}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getPackageId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setPackageId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bytes local = 3;
 * @return {string}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getLocal = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * optional bytes local = 3;
 * This is a type-conversion wrapper around `getLocal()`
 * @return {string}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getLocal_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getLocal()));
};


/**
 * optional bytes local = 3;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getLocal()`
 * @return {!Uint8Array}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getLocal_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getLocal()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setLocal = function(value) {
  return jspb.Message.setOneofField(this, 3, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_[0], value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.clearLocal = function() {
  return jspb.Message.setOneofField(this, 3, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_[0], undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.hasLocal = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional bool remote = 4;
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getRemote = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setRemote = function(value) {
  return jspb.Message.setOneofField(this, 4, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_[0], value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.clearRemote = function() {
  return jspb.Message.setOneofField(this, 4, proto.api_container_api.RunStarlarkPackageArgs.oneofGroups_[0], undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.hasRemote = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional string serialized_params = 5;
 * @return {string}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getSerializedParams = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setSerializedParams = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional bool dry_run = 6;
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getDryRun = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setDryRun = function(value) {
  return jspb.Message.setField(this, 6, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.clearDryRun = function() {
  return jspb.Message.setField(this, 6, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.hasDryRun = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional int32 parallelism = 7;
 * @return {number}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getParallelism = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setParallelism = function(value) {
  return jspb.Message.setField(this, 7, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.clearParallelism = function() {
  return jspb.Message.setField(this, 7, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.hasParallelism = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional bool clone_package = 8;
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.getClonePackage = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.setClonePackage = function(value) {
  return jspb.Message.setField(this, 8, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.RunStarlarkPackageArgs} returns this
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.clearClonePackage = function() {
  return jspb.Message.setField(this, 8, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RunStarlarkPackageArgs.prototype.hasClonePackage = function() {
  return jspb.Message.getField(this, 8) != null;
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.api_container_api.StarlarkRunResponseLine.oneofGroups_ = [[1,2,3,4,5,6]];

/**
 * @enum {number}
 */
proto.api_container_api.StarlarkRunResponseLine.RunResponseLineCase = {
  RUN_RESPONSE_LINE_NOT_SET: 0,
  INSTRUCTION: 1,
  ERROR: 2,
  PROGRESS_INFO: 3,
  INSTRUCTION_RESULT: 4,
  RUN_FINISHED_EVENT: 5,
  WARNING: 6
};

/**
 * @return {proto.api_container_api.StarlarkRunResponseLine.RunResponseLineCase}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getRunResponseLineCase = function() {
  return /** @type {proto.api_container_api.StarlarkRunResponseLine.RunResponseLineCase} */(jspb.Message.computeOneofCase(this, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0]));
};



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkRunResponseLine.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkRunResponseLine} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunResponseLine.toObject = function(includeInstance, msg) {
  var f, obj = {
    instruction: (f = msg.getInstruction()) && proto.api_container_api.StarlarkInstruction.toObject(includeInstance, f),
    error: (f = msg.getError()) && proto.api_container_api.StarlarkError.toObject(includeInstance, f),
    progressInfo: (f = msg.getProgressInfo()) && proto.api_container_api.StarlarkRunProgress.toObject(includeInstance, f),
    instructionResult: (f = msg.getInstructionResult()) && proto.api_container_api.StarlarkInstructionResult.toObject(includeInstance, f),
    runFinishedEvent: (f = msg.getRunFinishedEvent()) && proto.api_container_api.StarlarkRunFinishedEvent.toObject(includeInstance, f),
    warning: (f = msg.getWarning()) && proto.api_container_api.StarlarkWarning.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkRunResponseLine}
 */
proto.api_container_api.StarlarkRunResponseLine.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkRunResponseLine;
  return proto.api_container_api.StarlarkRunResponseLine.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkRunResponseLine} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkRunResponseLine}
 */
proto.api_container_api.StarlarkRunResponseLine.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api_container_api.StarlarkInstruction;
      reader.readMessage(value,proto.api_container_api.StarlarkInstruction.deserializeBinaryFromReader);
      msg.setInstruction(value);
      break;
    case 2:
      var value = new proto.api_container_api.StarlarkError;
      reader.readMessage(value,proto.api_container_api.StarlarkError.deserializeBinaryFromReader);
      msg.setError(value);
      break;
    case 3:
      var value = new proto.api_container_api.StarlarkRunProgress;
      reader.readMessage(value,proto.api_container_api.StarlarkRunProgress.deserializeBinaryFromReader);
      msg.setProgressInfo(value);
      break;
    case 4:
      var value = new proto.api_container_api.StarlarkInstructionResult;
      reader.readMessage(value,proto.api_container_api.StarlarkInstructionResult.deserializeBinaryFromReader);
      msg.setInstructionResult(value);
      break;
    case 5:
      var value = new proto.api_container_api.StarlarkRunFinishedEvent;
      reader.readMessage(value,proto.api_container_api.StarlarkRunFinishedEvent.deserializeBinaryFromReader);
      msg.setRunFinishedEvent(value);
      break;
    case 6:
      var value = new proto.api_container_api.StarlarkWarning;
      reader.readMessage(value,proto.api_container_api.StarlarkWarning.deserializeBinaryFromReader);
      msg.setWarning(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkRunResponseLine.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkRunResponseLine} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunResponseLine.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getInstruction();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api_container_api.StarlarkInstruction.serializeBinaryToWriter
    );
  }
  f = message.getError();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.api_container_api.StarlarkError.serializeBinaryToWriter
    );
  }
  f = message.getProgressInfo();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.api_container_api.StarlarkRunProgress.serializeBinaryToWriter
    );
  }
  f = message.getInstructionResult();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.api_container_api.StarlarkInstructionResult.serializeBinaryToWriter
    );
  }
  f = message.getRunFinishedEvent();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.api_container_api.StarlarkRunFinishedEvent.serializeBinaryToWriter
    );
  }
  f = message.getWarning();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.api_container_api.StarlarkWarning.serializeBinaryToWriter
    );
  }
};


/**
 * optional StarlarkInstruction instruction = 1;
 * @return {?proto.api_container_api.StarlarkInstruction}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getInstruction = function() {
  return /** @type{?proto.api_container_api.StarlarkInstruction} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkInstruction, 1));
};


/**
 * @param {?proto.api_container_api.StarlarkInstruction|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setInstruction = function(value) {
  return jspb.Message.setOneofWrapperField(this, 1, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearInstruction = function() {
  return this.setInstruction(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasInstruction = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional StarlarkError error = 2;
 * @return {?proto.api_container_api.StarlarkError}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getError = function() {
  return /** @type{?proto.api_container_api.StarlarkError} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkError, 2));
};


/**
 * @param {?proto.api_container_api.StarlarkError|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setError = function(value) {
  return jspb.Message.setOneofWrapperField(this, 2, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearError = function() {
  return this.setError(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasError = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional StarlarkRunProgress progress_info = 3;
 * @return {?proto.api_container_api.StarlarkRunProgress}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getProgressInfo = function() {
  return /** @type{?proto.api_container_api.StarlarkRunProgress} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkRunProgress, 3));
};


/**
 * @param {?proto.api_container_api.StarlarkRunProgress|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setProgressInfo = function(value) {
  return jspb.Message.setOneofWrapperField(this, 3, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearProgressInfo = function() {
  return this.setProgressInfo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasProgressInfo = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional StarlarkInstructionResult instruction_result = 4;
 * @return {?proto.api_container_api.StarlarkInstructionResult}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getInstructionResult = function() {
  return /** @type{?proto.api_container_api.StarlarkInstructionResult} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkInstructionResult, 4));
};


/**
 * @param {?proto.api_container_api.StarlarkInstructionResult|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setInstructionResult = function(value) {
  return jspb.Message.setOneofWrapperField(this, 4, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearInstructionResult = function() {
  return this.setInstructionResult(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasInstructionResult = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional StarlarkRunFinishedEvent run_finished_event = 5;
 * @return {?proto.api_container_api.StarlarkRunFinishedEvent}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getRunFinishedEvent = function() {
  return /** @type{?proto.api_container_api.StarlarkRunFinishedEvent} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkRunFinishedEvent, 5));
};


/**
 * @param {?proto.api_container_api.StarlarkRunFinishedEvent|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setRunFinishedEvent = function(value) {
  return jspb.Message.setOneofWrapperField(this, 5, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearRunFinishedEvent = function() {
  return this.setRunFinishedEvent(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasRunFinishedEvent = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional StarlarkWarning warning = 6;
 * @return {?proto.api_container_api.StarlarkWarning}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.getWarning = function() {
  return /** @type{?proto.api_container_api.StarlarkWarning} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkWarning, 6));
};


/**
 * @param {?proto.api_container_api.StarlarkWarning|undefined} value
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
*/
proto.api_container_api.StarlarkRunResponseLine.prototype.setWarning = function(value) {
  return jspb.Message.setOneofWrapperField(this, 6, proto.api_container_api.StarlarkRunResponseLine.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunResponseLine} returns this
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.clearWarning = function() {
  return this.setWarning(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunResponseLine.prototype.hasWarning = function() {
  return jspb.Message.getField(this, 6) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkWarning.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkWarning.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkWarning} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkWarning.toObject = function(includeInstance, msg) {
  var f, obj = {
    warningMessage: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkWarning}
 */
proto.api_container_api.StarlarkWarning.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkWarning;
  return proto.api_container_api.StarlarkWarning.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkWarning} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkWarning}
 */
proto.api_container_api.StarlarkWarning.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setWarningMessage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkWarning.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkWarning.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkWarning} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkWarning.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getWarningMessage();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string warning_message = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkWarning.prototype.getWarningMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkWarning} returns this
 */
proto.api_container_api.StarlarkWarning.prototype.setWarningMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.StarlarkInstruction.repeatedFields_ = [3];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkInstruction.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkInstruction.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkInstruction} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstruction.toObject = function(includeInstance, msg) {
  var f, obj = {
    position: (f = msg.getPosition()) && proto.api_container_api.StarlarkInstructionPosition.toObject(includeInstance, f),
    instructionName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    argumentsList: jspb.Message.toObjectList(msg.getArgumentsList(),
    proto.api_container_api.StarlarkInstructionArg.toObject, includeInstance),
    executableInstruction: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkInstruction}
 */
proto.api_container_api.StarlarkInstruction.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkInstruction;
  return proto.api_container_api.StarlarkInstruction.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkInstruction} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkInstruction}
 */
proto.api_container_api.StarlarkInstruction.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api_container_api.StarlarkInstructionPosition;
      reader.readMessage(value,proto.api_container_api.StarlarkInstructionPosition.deserializeBinaryFromReader);
      msg.setPosition(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setInstructionName(value);
      break;
    case 3:
      var value = new proto.api_container_api.StarlarkInstructionArg;
      reader.readMessage(value,proto.api_container_api.StarlarkInstructionArg.deserializeBinaryFromReader);
      msg.addArguments(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setExecutableInstruction(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkInstruction.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkInstruction.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkInstruction} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstruction.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPosition();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api_container_api.StarlarkInstructionPosition.serializeBinaryToWriter
    );
  }
  f = message.getInstructionName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getArgumentsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.api_container_api.StarlarkInstructionArg.serializeBinaryToWriter
    );
  }
  f = message.getExecutableInstruction();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional StarlarkInstructionPosition position = 1;
 * @return {?proto.api_container_api.StarlarkInstructionPosition}
 */
proto.api_container_api.StarlarkInstruction.prototype.getPosition = function() {
  return /** @type{?proto.api_container_api.StarlarkInstructionPosition} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkInstructionPosition, 1));
};


/**
 * @param {?proto.api_container_api.StarlarkInstructionPosition|undefined} value
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
*/
proto.api_container_api.StarlarkInstruction.prototype.setPosition = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
 */
proto.api_container_api.StarlarkInstruction.prototype.clearPosition = function() {
  return this.setPosition(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkInstruction.prototype.hasPosition = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string instruction_name = 2;
 * @return {string}
 */
proto.api_container_api.StarlarkInstruction.prototype.getInstructionName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
 */
proto.api_container_api.StarlarkInstruction.prototype.setInstructionName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated StarlarkInstructionArg arguments = 3;
 * @return {!Array<!proto.api_container_api.StarlarkInstructionArg>}
 */
proto.api_container_api.StarlarkInstruction.prototype.getArgumentsList = function() {
  return /** @type{!Array<!proto.api_container_api.StarlarkInstructionArg>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api_container_api.StarlarkInstructionArg, 3));
};


/**
 * @param {!Array<!proto.api_container_api.StarlarkInstructionArg>} value
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
*/
proto.api_container_api.StarlarkInstruction.prototype.setArgumentsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.api_container_api.StarlarkInstructionArg=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.StarlarkInstructionArg}
 */
proto.api_container_api.StarlarkInstruction.prototype.addArguments = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.api_container_api.StarlarkInstructionArg, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
 */
proto.api_container_api.StarlarkInstruction.prototype.clearArgumentsList = function() {
  return this.setArgumentsList([]);
};


/**
 * optional string executable_instruction = 4;
 * @return {string}
 */
proto.api_container_api.StarlarkInstruction.prototype.getExecutableInstruction = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstruction} returns this
 */
proto.api_container_api.StarlarkInstruction.prototype.setExecutableInstruction = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkInstructionResult.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkInstructionResult.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkInstructionResult} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionResult.toObject = function(includeInstance, msg) {
  var f, obj = {
    serializedInstructionResult: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkInstructionResult}
 */
proto.api_container_api.StarlarkInstructionResult.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkInstructionResult;
  return proto.api_container_api.StarlarkInstructionResult.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkInstructionResult} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkInstructionResult}
 */
proto.api_container_api.StarlarkInstructionResult.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedInstructionResult(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkInstructionResult.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkInstructionResult.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkInstructionResult} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionResult.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSerializedInstructionResult();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string serialized_instruction_result = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkInstructionResult.prototype.getSerializedInstructionResult = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstructionResult} returns this
 */
proto.api_container_api.StarlarkInstructionResult.prototype.setSerializedInstructionResult = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkInstructionArg.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkInstructionArg} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionArg.toObject = function(includeInstance, msg) {
  var f, obj = {
    serializedArgValue: jspb.Message.getFieldWithDefault(msg, 1, ""),
    argName: jspb.Message.getFieldWithDefault(msg, 2, ""),
    isRepresentative: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkInstructionArg}
 */
proto.api_container_api.StarlarkInstructionArg.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkInstructionArg;
  return proto.api_container_api.StarlarkInstructionArg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkInstructionArg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkInstructionArg}
 */
proto.api_container_api.StarlarkInstructionArg.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedArgValue(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setArgName(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsRepresentative(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkInstructionArg.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkInstructionArg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionArg.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSerializedArgValue();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 2));
  if (f != null) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getIsRepresentative();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string serialized_arg_value = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.getSerializedArgValue = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstructionArg} returns this
 */
proto.api_container_api.StarlarkInstructionArg.prototype.setSerializedArgValue = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string arg_name = 2;
 * @return {string}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.getArgName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstructionArg} returns this
 */
proto.api_container_api.StarlarkInstructionArg.prototype.setArgName = function(value) {
  return jspb.Message.setField(this, 2, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.StarlarkInstructionArg} returns this
 */
proto.api_container_api.StarlarkInstructionArg.prototype.clearArgName = function() {
  return jspb.Message.setField(this, 2, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.hasArgName = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool is_representative = 3;
 * @return {boolean}
 */
proto.api_container_api.StarlarkInstructionArg.prototype.getIsRepresentative = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.StarlarkInstructionArg} returns this
 */
proto.api_container_api.StarlarkInstructionArg.prototype.setIsRepresentative = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkInstructionPosition.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkInstructionPosition} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionPosition.toObject = function(includeInstance, msg) {
  var f, obj = {
    filename: jspb.Message.getFieldWithDefault(msg, 1, ""),
    line: jspb.Message.getFieldWithDefault(msg, 2, 0),
    column: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkInstructionPosition}
 */
proto.api_container_api.StarlarkInstructionPosition.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkInstructionPosition;
  return proto.api_container_api.StarlarkInstructionPosition.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkInstructionPosition} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkInstructionPosition}
 */
proto.api_container_api.StarlarkInstructionPosition.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilename(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setLine(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setColumn(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkInstructionPosition.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkInstructionPosition} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInstructionPosition.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFilename();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getLine();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getColumn();
  if (f !== 0) {
    writer.writeInt32(
      3,
      f
    );
  }
};


/**
 * optional string filename = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.getFilename = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInstructionPosition} returns this
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.setFilename = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int32 line = 2;
 * @return {number}
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.getLine = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.StarlarkInstructionPosition} returns this
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.setLine = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int32 column = 3;
 * @return {number}
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.getColumn = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.StarlarkInstructionPosition} returns this
 */
proto.api_container_api.StarlarkInstructionPosition.prototype.setColumn = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.api_container_api.StarlarkError.oneofGroups_ = [[1,2,3]];

/**
 * @enum {number}
 */
proto.api_container_api.StarlarkError.ErrorCase = {
  ERROR_NOT_SET: 0,
  INTERPRETATION_ERROR: 1,
  VALIDATION_ERROR: 2,
  EXECUTION_ERROR: 3
};

/**
 * @return {proto.api_container_api.StarlarkError.ErrorCase}
 */
proto.api_container_api.StarlarkError.prototype.getErrorCase = function() {
  return /** @type {proto.api_container_api.StarlarkError.ErrorCase} */(jspb.Message.computeOneofCase(this, proto.api_container_api.StarlarkError.oneofGroups_[0]));
};



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkError.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkError.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkError} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkError.toObject = function(includeInstance, msg) {
  var f, obj = {
    interpretationError: (f = msg.getInterpretationError()) && proto.api_container_api.StarlarkInterpretationError.toObject(includeInstance, f),
    validationError: (f = msg.getValidationError()) && proto.api_container_api.StarlarkValidationError.toObject(includeInstance, f),
    executionError: (f = msg.getExecutionError()) && proto.api_container_api.StarlarkExecutionError.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkError}
 */
proto.api_container_api.StarlarkError.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkError;
  return proto.api_container_api.StarlarkError.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkError} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkError}
 */
proto.api_container_api.StarlarkError.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api_container_api.StarlarkInterpretationError;
      reader.readMessage(value,proto.api_container_api.StarlarkInterpretationError.deserializeBinaryFromReader);
      msg.setInterpretationError(value);
      break;
    case 2:
      var value = new proto.api_container_api.StarlarkValidationError;
      reader.readMessage(value,proto.api_container_api.StarlarkValidationError.deserializeBinaryFromReader);
      msg.setValidationError(value);
      break;
    case 3:
      var value = new proto.api_container_api.StarlarkExecutionError;
      reader.readMessage(value,proto.api_container_api.StarlarkExecutionError.deserializeBinaryFromReader);
      msg.setExecutionError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkError.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkError.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkError} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkError.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getInterpretationError();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.api_container_api.StarlarkInterpretationError.serializeBinaryToWriter
    );
  }
  f = message.getValidationError();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.api_container_api.StarlarkValidationError.serializeBinaryToWriter
    );
  }
  f = message.getExecutionError();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.api_container_api.StarlarkExecutionError.serializeBinaryToWriter
    );
  }
};


/**
 * optional StarlarkInterpretationError interpretation_error = 1;
 * @return {?proto.api_container_api.StarlarkInterpretationError}
 */
proto.api_container_api.StarlarkError.prototype.getInterpretationError = function() {
  return /** @type{?proto.api_container_api.StarlarkInterpretationError} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkInterpretationError, 1));
};


/**
 * @param {?proto.api_container_api.StarlarkInterpretationError|undefined} value
 * @return {!proto.api_container_api.StarlarkError} returns this
*/
proto.api_container_api.StarlarkError.prototype.setInterpretationError = function(value) {
  return jspb.Message.setOneofWrapperField(this, 1, proto.api_container_api.StarlarkError.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkError} returns this
 */
proto.api_container_api.StarlarkError.prototype.clearInterpretationError = function() {
  return this.setInterpretationError(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkError.prototype.hasInterpretationError = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional StarlarkValidationError validation_error = 2;
 * @return {?proto.api_container_api.StarlarkValidationError}
 */
proto.api_container_api.StarlarkError.prototype.getValidationError = function() {
  return /** @type{?proto.api_container_api.StarlarkValidationError} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkValidationError, 2));
};


/**
 * @param {?proto.api_container_api.StarlarkValidationError|undefined} value
 * @return {!proto.api_container_api.StarlarkError} returns this
*/
proto.api_container_api.StarlarkError.prototype.setValidationError = function(value) {
  return jspb.Message.setOneofWrapperField(this, 2, proto.api_container_api.StarlarkError.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkError} returns this
 */
proto.api_container_api.StarlarkError.prototype.clearValidationError = function() {
  return this.setValidationError(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkError.prototype.hasValidationError = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional StarlarkExecutionError execution_error = 3;
 * @return {?proto.api_container_api.StarlarkExecutionError}
 */
proto.api_container_api.StarlarkError.prototype.getExecutionError = function() {
  return /** @type{?proto.api_container_api.StarlarkExecutionError} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.StarlarkExecutionError, 3));
};


/**
 * @param {?proto.api_container_api.StarlarkExecutionError|undefined} value
 * @return {!proto.api_container_api.StarlarkError} returns this
*/
proto.api_container_api.StarlarkError.prototype.setExecutionError = function(value) {
  return jspb.Message.setOneofWrapperField(this, 3, proto.api_container_api.StarlarkError.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StarlarkError} returns this
 */
proto.api_container_api.StarlarkError.prototype.clearExecutionError = function() {
  return this.setExecutionError(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkError.prototype.hasExecutionError = function() {
  return jspb.Message.getField(this, 3) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkInterpretationError.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkInterpretationError.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkInterpretationError} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInterpretationError.toObject = function(includeInstance, msg) {
  var f, obj = {
    errorMessage: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkInterpretationError}
 */
proto.api_container_api.StarlarkInterpretationError.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkInterpretationError;
  return proto.api_container_api.StarlarkInterpretationError.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkInterpretationError} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkInterpretationError}
 */
proto.api_container_api.StarlarkInterpretationError.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setErrorMessage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkInterpretationError.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkInterpretationError.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkInterpretationError} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkInterpretationError.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getErrorMessage();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string error_message = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkInterpretationError.prototype.getErrorMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkInterpretationError} returns this
 */
proto.api_container_api.StarlarkInterpretationError.prototype.setErrorMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkValidationError.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkValidationError.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkValidationError} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkValidationError.toObject = function(includeInstance, msg) {
  var f, obj = {
    errorMessage: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkValidationError}
 */
proto.api_container_api.StarlarkValidationError.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkValidationError;
  return proto.api_container_api.StarlarkValidationError.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkValidationError} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkValidationError}
 */
proto.api_container_api.StarlarkValidationError.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setErrorMessage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkValidationError.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkValidationError.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkValidationError} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkValidationError.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getErrorMessage();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string error_message = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkValidationError.prototype.getErrorMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkValidationError} returns this
 */
proto.api_container_api.StarlarkValidationError.prototype.setErrorMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkExecutionError.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkExecutionError.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkExecutionError} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkExecutionError.toObject = function(includeInstance, msg) {
  var f, obj = {
    errorMessage: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkExecutionError}
 */
proto.api_container_api.StarlarkExecutionError.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkExecutionError;
  return proto.api_container_api.StarlarkExecutionError.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkExecutionError} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkExecutionError}
 */
proto.api_container_api.StarlarkExecutionError.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setErrorMessage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkExecutionError.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkExecutionError.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkExecutionError} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkExecutionError.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getErrorMessage();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string error_message = 1;
 * @return {string}
 */
proto.api_container_api.StarlarkExecutionError.prototype.getErrorMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkExecutionError} returns this
 */
proto.api_container_api.StarlarkExecutionError.prototype.setErrorMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.StarlarkRunProgress.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkRunProgress.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkRunProgress.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkRunProgress} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunProgress.toObject = function(includeInstance, msg) {
  var f, obj = {
    currentStepInfoList: (f = jspb.Message.getRepeatedField(msg, 1)) == null ? undefined : f,
    totalSteps: jspb.Message.getFieldWithDefault(msg, 2, 0),
    currentStepNumber: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkRunProgress}
 */
proto.api_container_api.StarlarkRunProgress.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkRunProgress;
  return proto.api_container_api.StarlarkRunProgress.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkRunProgress} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkRunProgress}
 */
proto.api_container_api.StarlarkRunProgress.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.addCurrentStepInfo(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setTotalSteps(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setCurrentStepNumber(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkRunProgress.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkRunProgress.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkRunProgress} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunProgress.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCurrentStepInfoList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      1,
      f
    );
  }
  f = message.getTotalSteps();
  if (f !== 0) {
    writer.writeUint32(
      2,
      f
    );
  }
  f = message.getCurrentStepNumber();
  if (f !== 0) {
    writer.writeUint32(
      3,
      f
    );
  }
};


/**
 * repeated string current_step_info = 1;
 * @return {!Array<string>}
 */
proto.api_container_api.StarlarkRunProgress.prototype.getCurrentStepInfoList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 1));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api_container_api.StarlarkRunProgress} returns this
 */
proto.api_container_api.StarlarkRunProgress.prototype.setCurrentStepInfoList = function(value) {
  return jspb.Message.setField(this, 1, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.StarlarkRunProgress} returns this
 */
proto.api_container_api.StarlarkRunProgress.prototype.addCurrentStepInfo = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 1, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.StarlarkRunProgress} returns this
 */
proto.api_container_api.StarlarkRunProgress.prototype.clearCurrentStepInfoList = function() {
  return this.setCurrentStepInfoList([]);
};


/**
 * optional uint32 total_steps = 2;
 * @return {number}
 */
proto.api_container_api.StarlarkRunProgress.prototype.getTotalSteps = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.StarlarkRunProgress} returns this
 */
proto.api_container_api.StarlarkRunProgress.prototype.setTotalSteps = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional uint32 current_step_number = 3;
 * @return {number}
 */
proto.api_container_api.StarlarkRunProgress.prototype.getCurrentStepNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.StarlarkRunProgress} returns this
 */
proto.api_container_api.StarlarkRunProgress.prototype.setCurrentStepNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StarlarkRunFinishedEvent.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StarlarkRunFinishedEvent} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunFinishedEvent.toObject = function(includeInstance, msg) {
  var f, obj = {
    isrunsuccessful: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    serializedOutput: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StarlarkRunFinishedEvent}
 */
proto.api_container_api.StarlarkRunFinishedEvent.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StarlarkRunFinishedEvent;
  return proto.api_container_api.StarlarkRunFinishedEvent.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StarlarkRunFinishedEvent} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StarlarkRunFinishedEvent}
 */
proto.api_container_api.StarlarkRunFinishedEvent.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIsrunsuccessful(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSerializedOutput(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StarlarkRunFinishedEvent.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StarlarkRunFinishedEvent} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StarlarkRunFinishedEvent.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getIsrunsuccessful();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = /** @type {string} */ (jspb.Message.getField(message, 2));
  if (f != null) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional bool isRunSuccessful = 1;
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.getIsrunsuccessful = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.api_container_api.StarlarkRunFinishedEvent} returns this
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.setIsrunsuccessful = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * optional string serialized_output = 2;
 * @return {string}
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.getSerializedOutput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StarlarkRunFinishedEvent} returns this
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.setSerializedOutput = function(value) {
  return jspb.Message.setField(this, 2, value);
};


/**
 * Clears the field making it undefined.
 * @return {!proto.api_container_api.StarlarkRunFinishedEvent} returns this
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.clearSerializedOutput = function() {
  return jspb.Message.setField(this, 2, undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StarlarkRunFinishedEvent.prototype.hasSerializedOutput = function() {
  return jspb.Message.getField(this, 2) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StartServicesArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StartServicesArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StartServicesArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StartServicesArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceNamesToConfigsMap: (f = msg.getServiceNamesToConfigsMap()) ? f.toObject(includeInstance, proto.api_container_api.ServiceConfig.toObject) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StartServicesArgs}
 */
proto.api_container_api.StartServicesArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StartServicesArgs;
  return proto.api_container_api.StartServicesArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StartServicesArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StartServicesArgs}
 */
proto.api_container_api.StartServicesArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getServiceNamesToConfigsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.ServiceConfig.deserializeBinaryFromReader, "", new proto.api_container_api.ServiceConfig());
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StartServicesArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StartServicesArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StartServicesArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StartServicesArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceNamesToConfigsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.ServiceConfig.serializeBinaryToWriter);
  }
};


/**
 * map<string, ServiceConfig> service_names_to_configs = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.ServiceConfig>}
 */
proto.api_container_api.StartServicesArgs.prototype.getServiceNamesToConfigsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.ServiceConfig>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.ServiceConfig));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.StartServicesArgs} returns this
 */
proto.api_container_api.StartServicesArgs.prototype.clearServiceNamesToConfigsMap = function() {
  this.getServiceNamesToConfigsMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StartServicesResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StartServicesResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StartServicesResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StartServicesResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    successfulServiceNameToServiceInfoMap: (f = msg.getSuccessfulServiceNameToServiceInfoMap()) ? f.toObject(includeInstance, proto.api_container_api.ServiceInfo.toObject) : [],
    failedServiceNameToErrorMap: (f = msg.getFailedServiceNameToErrorMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StartServicesResponse}
 */
proto.api_container_api.StartServicesResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StartServicesResponse;
  return proto.api_container_api.StartServicesResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StartServicesResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StartServicesResponse}
 */
proto.api_container_api.StartServicesResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getSuccessfulServiceNameToServiceInfoMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.ServiceInfo.deserializeBinaryFromReader, "", new proto.api_container_api.ServiceInfo());
         });
      break;
    case 2:
      var value = msg.getFailedServiceNameToErrorMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StartServicesResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StartServicesResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StartServicesResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StartServicesResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSuccessfulServiceNameToServiceInfoMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.ServiceInfo.serializeBinaryToWriter);
  }
  f = message.getFailedServiceNameToErrorMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * map<string, ServiceInfo> successful_service_name_to_service_info = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.ServiceInfo>}
 */
proto.api_container_api.StartServicesResponse.prototype.getSuccessfulServiceNameToServiceInfoMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.ServiceInfo>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.ServiceInfo));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.StartServicesResponse} returns this
 */
proto.api_container_api.StartServicesResponse.prototype.clearSuccessfulServiceNameToServiceInfoMap = function() {
  this.getSuccessfulServiceNameToServiceInfoMap().clear();
  return this;};


/**
 * map<string, string> failed_service_name_to_error = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.api_container_api.StartServicesResponse.prototype.getFailedServiceNameToErrorMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.StartServicesResponse} returns this
 */
proto.api_container_api.StartServicesResponse.prototype.clearFailedServiceNameToErrorMap = function() {
  this.getFailedServiceNameToErrorMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.GetServicesArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.GetServicesArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.GetServicesArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetServicesArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifiersMap: (f = msg.getServiceIdentifiersMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.GetServicesArgs}
 */
proto.api_container_api.GetServicesArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.GetServicesArgs;
  return proto.api_container_api.GetServicesArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.GetServicesArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.GetServicesArgs}
 */
proto.api_container_api.GetServicesArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getServiceIdentifiersMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readBool, null, "", false);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.GetServicesArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.GetServicesArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.GetServicesArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetServicesArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifiersMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeBool);
  }
};


/**
 * map<string, bool> service_identifiers = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,boolean>}
 */
proto.api_container_api.GetServicesArgs.prototype.getServiceIdentifiersMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,boolean>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.GetServicesArgs} returns this
 */
proto.api_container_api.GetServicesArgs.prototype.clearServiceIdentifiersMap = function() {
  this.getServiceIdentifiersMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.GetServicesResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.GetServicesResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.GetServicesResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetServicesResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceInfoMap: (f = msg.getServiceInfoMap()) ? f.toObject(includeInstance, proto.api_container_api.ServiceInfo.toObject) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.GetServicesResponse}
 */
proto.api_container_api.GetServicesResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.GetServicesResponse;
  return proto.api_container_api.GetServicesResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.GetServicesResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.GetServicesResponse}
 */
proto.api_container_api.GetServicesResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getServiceInfoMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.ServiceInfo.deserializeBinaryFromReader, "", new proto.api_container_api.ServiceInfo());
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.GetServicesResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.GetServicesResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.GetServicesResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetServicesResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceInfoMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.ServiceInfo.serializeBinaryToWriter);
  }
};


/**
 * map<string, ServiceInfo> service_info = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.ServiceInfo>}
 */
proto.api_container_api.GetServicesResponse.prototype.getServiceInfoMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.ServiceInfo>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.ServiceInfo));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.GetServicesResponse} returns this
 */
proto.api_container_api.GetServicesResponse.prototype.clearServiceInfoMap = function() {
  this.getServiceInfoMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ServiceIdentifiers.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ServiceIdentifiers.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ServiceIdentifiers} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceIdentifiers.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceUuid: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, ""),
    shortenedUuid: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ServiceIdentifiers}
 */
proto.api_container_api.ServiceIdentifiers.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ServiceIdentifiers;
  return proto.api_container_api.ServiceIdentifiers.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ServiceIdentifiers} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ServiceIdentifiers}
 */
proto.api_container_api.ServiceIdentifiers.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceUuid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setShortenedUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ServiceIdentifiers.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ServiceIdentifiers.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ServiceIdentifiers} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ServiceIdentifiers.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getShortenedUuid();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string service_uuid = 1;
 * @return {string}
 */
proto.api_container_api.ServiceIdentifiers.prototype.getServiceUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceIdentifiers} returns this
 */
proto.api_container_api.ServiceIdentifiers.prototype.setServiceUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api_container_api.ServiceIdentifiers.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceIdentifiers} returns this
 */
proto.api_container_api.ServiceIdentifiers.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string shortened_uuid = 3;
 * @return {string}
 */
proto.api_container_api.ServiceIdentifiers.prototype.getShortenedUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ServiceIdentifiers} returns this
 */
proto.api_container_api.ServiceIdentifiers.prototype.setShortenedUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    allidentifiersList: jspb.Message.toObjectList(msg.getAllidentifiersList(),
    proto.api_container_api.ServiceIdentifiers.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse;
  return proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api_container_api.ServiceIdentifiers;
      reader.readMessage(value,proto.api_container_api.ServiceIdentifiers.deserializeBinaryFromReader);
      msg.addAllidentifiers(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAllidentifiersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api_container_api.ServiceIdentifiers.serializeBinaryToWriter
    );
  }
};


/**
 * repeated ServiceIdentifiers allIdentifiers = 1;
 * @return {!Array<!proto.api_container_api.ServiceIdentifiers>}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.getAllidentifiersList = function() {
  return /** @type{!Array<!proto.api_container_api.ServiceIdentifiers>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api_container_api.ServiceIdentifiers, 1));
};


/**
 * @param {!Array<!proto.api_container_api.ServiceIdentifiers>} value
 * @return {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse} returns this
*/
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.setAllidentifiersList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api_container_api.ServiceIdentifiers=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.ServiceIdentifiers}
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.addAllidentifiers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api_container_api.ServiceIdentifiers, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse} returns this
 */
proto.api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse.prototype.clearAllidentifiersList = function() {
  return this.setAllidentifiersList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RemoveServiceArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RemoveServiceArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RemoveServiceArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RemoveServiceArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RemoveServiceArgs}
 */
proto.api_container_api.RemoveServiceArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RemoveServiceArgs;
  return proto.api_container_api.RemoveServiceArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RemoveServiceArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RemoveServiceArgs}
 */
proto.api_container_api.RemoveServiceArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RemoveServiceArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RemoveServiceArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RemoveServiceArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RemoveServiceArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.RemoveServiceArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RemoveServiceArgs} returns this
 */
proto.api_container_api.RemoveServiceArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RemoveServiceResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RemoveServiceResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RemoveServiceResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RemoveServiceResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceUuid: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RemoveServiceResponse}
 */
proto.api_container_api.RemoveServiceResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RemoveServiceResponse;
  return proto.api_container_api.RemoveServiceResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RemoveServiceResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RemoveServiceResponse}
 */
proto.api_container_api.RemoveServiceResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RemoveServiceResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RemoveServiceResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RemoveServiceResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RemoveServiceResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string service_uuid = 1;
 * @return {string}
 */
proto.api_container_api.RemoveServiceResponse.prototype.getServiceUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RemoveServiceResponse} returns this
 */
proto.api_container_api.RemoveServiceResponse.prototype.setServiceUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RepartitionArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RepartitionArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RepartitionArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RepartitionArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    partitionServicesMap: (f = msg.getPartitionServicesMap()) ? f.toObject(includeInstance, proto.api_container_api.PartitionServices.toObject) : [],
    partitionConnectionsMap: (f = msg.getPartitionConnectionsMap()) ? f.toObject(includeInstance, proto.api_container_api.PartitionConnections.toObject) : [],
    defaultConnection: (f = msg.getDefaultConnection()) && proto.api_container_api.PartitionConnectionInfo.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RepartitionArgs}
 */
proto.api_container_api.RepartitionArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RepartitionArgs;
  return proto.api_container_api.RepartitionArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RepartitionArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RepartitionArgs}
 */
proto.api_container_api.RepartitionArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getPartitionServicesMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.PartitionServices.deserializeBinaryFromReader, "", new proto.api_container_api.PartitionServices());
         });
      break;
    case 2:
      var value = msg.getPartitionConnectionsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.PartitionConnections.deserializeBinaryFromReader, "", new proto.api_container_api.PartitionConnections());
         });
      break;
    case 3:
      var value = new proto.api_container_api.PartitionConnectionInfo;
      reader.readMessage(value,proto.api_container_api.PartitionConnectionInfo.deserializeBinaryFromReader);
      msg.setDefaultConnection(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RepartitionArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RepartitionArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RepartitionArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RepartitionArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPartitionServicesMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.PartitionServices.serializeBinaryToWriter);
  }
  f = message.getPartitionConnectionsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.PartitionConnections.serializeBinaryToWriter);
  }
  f = message.getDefaultConnection();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.api_container_api.PartitionConnectionInfo.serializeBinaryToWriter
    );
  }
};


/**
 * map<string, PartitionServices> partition_services = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.PartitionServices>}
 */
proto.api_container_api.RepartitionArgs.prototype.getPartitionServicesMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.PartitionServices>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.PartitionServices));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.RepartitionArgs} returns this
 */
proto.api_container_api.RepartitionArgs.prototype.clearPartitionServicesMap = function() {
  this.getPartitionServicesMap().clear();
  return this;};


/**
 * map<string, PartitionConnections> partition_connections = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.PartitionConnections>}
 */
proto.api_container_api.RepartitionArgs.prototype.getPartitionConnectionsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.PartitionConnections>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      proto.api_container_api.PartitionConnections));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.RepartitionArgs} returns this
 */
proto.api_container_api.RepartitionArgs.prototype.clearPartitionConnectionsMap = function() {
  this.getPartitionConnectionsMap().clear();
  return this;};


/**
 * optional PartitionConnectionInfo default_connection = 3;
 * @return {?proto.api_container_api.PartitionConnectionInfo}
 */
proto.api_container_api.RepartitionArgs.prototype.getDefaultConnection = function() {
  return /** @type{?proto.api_container_api.PartitionConnectionInfo} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.PartitionConnectionInfo, 3));
};


/**
 * @param {?proto.api_container_api.PartitionConnectionInfo|undefined} value
 * @return {!proto.api_container_api.RepartitionArgs} returns this
*/
proto.api_container_api.RepartitionArgs.prototype.setDefaultConnection = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.RepartitionArgs} returns this
 */
proto.api_container_api.RepartitionArgs.prototype.clearDefaultConnection = function() {
  return this.setDefaultConnection(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.RepartitionArgs.prototype.hasDefaultConnection = function() {
  return jspb.Message.getField(this, 3) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.PartitionServices.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.PartitionServices.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.PartitionServices} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionServices.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceNameSetMap: (f = msg.getServiceNameSetMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.PartitionServices}
 */
proto.api_container_api.PartitionServices.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.PartitionServices;
  return proto.api_container_api.PartitionServices.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.PartitionServices} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.PartitionServices}
 */
proto.api_container_api.PartitionServices.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getServiceNameSetMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readBool, null, "", false);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.PartitionServices.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.PartitionServices.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.PartitionServices} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionServices.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceNameSetMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeBool);
  }
};


/**
 * map<string, bool> service_name_set = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,boolean>}
 */
proto.api_container_api.PartitionServices.prototype.getServiceNameSetMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,boolean>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.PartitionServices} returns this
 */
proto.api_container_api.PartitionServices.prototype.clearServiceNameSetMap = function() {
  this.getServiceNameSetMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.PartitionConnections.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.PartitionConnections.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.PartitionConnections} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionConnections.toObject = function(includeInstance, msg) {
  var f, obj = {
    connectionInfoMap: (f = msg.getConnectionInfoMap()) ? f.toObject(includeInstance, proto.api_container_api.PartitionConnectionInfo.toObject) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.PartitionConnections}
 */
proto.api_container_api.PartitionConnections.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.PartitionConnections;
  return proto.api_container_api.PartitionConnections.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.PartitionConnections} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.PartitionConnections}
 */
proto.api_container_api.PartitionConnections.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getConnectionInfoMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.PartitionConnectionInfo.deserializeBinaryFromReader, "", new proto.api_container_api.PartitionConnectionInfo());
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.PartitionConnections.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.PartitionConnections.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.PartitionConnections} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionConnections.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getConnectionInfoMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.PartitionConnectionInfo.serializeBinaryToWriter);
  }
};


/**
 * map<string, PartitionConnectionInfo> connection_info = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.PartitionConnectionInfo>}
 */
proto.api_container_api.PartitionConnections.prototype.getConnectionInfoMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.PartitionConnectionInfo>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.PartitionConnectionInfo));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.PartitionConnections} returns this
 */
proto.api_container_api.PartitionConnections.prototype.clearConnectionInfoMap = function() {
  this.getConnectionInfoMap().clear();
  return this;};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.PartitionConnectionInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.PartitionConnectionInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.PartitionConnectionInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionConnectionInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    packetLossPercentage: jspb.Message.getFloatingPointFieldWithDefault(msg, 1, 0.0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.PartitionConnectionInfo}
 */
proto.api_container_api.PartitionConnectionInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.PartitionConnectionInfo;
  return proto.api_container_api.PartitionConnectionInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.PartitionConnectionInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.PartitionConnectionInfo}
 */
proto.api_container_api.PartitionConnectionInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readFloat());
      msg.setPacketLossPercentage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.PartitionConnectionInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.PartitionConnectionInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.PartitionConnectionInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PartitionConnectionInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPacketLossPercentage();
  if (f !== 0.0) {
    writer.writeFloat(
      1,
      f
    );
  }
};


/**
 * optional float packet_loss_percentage = 1;
 * @return {number}
 */
proto.api_container_api.PartitionConnectionInfo.prototype.getPacketLossPercentage = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 1, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.PartitionConnectionInfo} returns this
 */
proto.api_container_api.PartitionConnectionInfo.prototype.setPacketLossPercentage = function(value) {
  return jspb.Message.setProto3FloatField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.ExecCommandArgs.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ExecCommandArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ExecCommandArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ExecCommandArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ExecCommandArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, ""),
    commandArgsList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ExecCommandArgs}
 */
proto.api_container_api.ExecCommandArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ExecCommandArgs;
  return proto.api_container_api.ExecCommandArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ExecCommandArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ExecCommandArgs}
 */
proto.api_container_api.ExecCommandArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addCommandArgs(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ExecCommandArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ExecCommandArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ExecCommandArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ExecCommandArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getCommandArgsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.ExecCommandArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ExecCommandArgs} returns this
 */
proto.api_container_api.ExecCommandArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string command_args = 2;
 * @return {!Array<string>}
 */
proto.api_container_api.ExecCommandArgs.prototype.getCommandArgsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.api_container_api.ExecCommandArgs} returns this
 */
proto.api_container_api.ExecCommandArgs.prototype.setCommandArgsList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.ExecCommandArgs} returns this
 */
proto.api_container_api.ExecCommandArgs.prototype.addCommandArgs = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.ExecCommandArgs} returns this
 */
proto.api_container_api.ExecCommandArgs.prototype.clearCommandArgsList = function() {
  return this.setCommandArgsList([]);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.PauseServiceArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.PauseServiceArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.PauseServiceArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PauseServiceArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.PauseServiceArgs}
 */
proto.api_container_api.PauseServiceArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.PauseServiceArgs;
  return proto.api_container_api.PauseServiceArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.PauseServiceArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.PauseServiceArgs}
 */
proto.api_container_api.PauseServiceArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.PauseServiceArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.PauseServiceArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.PauseServiceArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.PauseServiceArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.PauseServiceArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.PauseServiceArgs} returns this
 */
proto.api_container_api.PauseServiceArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.UnpauseServiceArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.UnpauseServiceArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.UnpauseServiceArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UnpauseServiceArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.UnpauseServiceArgs}
 */
proto.api_container_api.UnpauseServiceArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.UnpauseServiceArgs;
  return proto.api_container_api.UnpauseServiceArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.UnpauseServiceArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.UnpauseServiceArgs}
 */
proto.api_container_api.UnpauseServiceArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.UnpauseServiceArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.UnpauseServiceArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.UnpauseServiceArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UnpauseServiceArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.UnpauseServiceArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.UnpauseServiceArgs} returns this
 */
proto.api_container_api.UnpauseServiceArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ExecCommandResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ExecCommandResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ExecCommandResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ExecCommandResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    exitCode: jspb.Message.getFieldWithDefault(msg, 1, 0),
    logOutput: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ExecCommandResponse}
 */
proto.api_container_api.ExecCommandResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ExecCommandResponse;
  return proto.api_container_api.ExecCommandResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ExecCommandResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ExecCommandResponse}
 */
proto.api_container_api.ExecCommandResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setExitCode(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setLogOutput(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ExecCommandResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ExecCommandResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ExecCommandResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ExecCommandResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getExitCode();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getLogOutput();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional int32 exit_code = 1;
 * @return {number}
 */
proto.api_container_api.ExecCommandResponse.prototype.getExitCode = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.ExecCommandResponse} returns this
 */
proto.api_container_api.ExecCommandResponse.prototype.setExitCode = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string log_output = 2;
 * @return {string}
 */
proto.api_container_api.ExecCommandResponse.prototype.getLogOutput = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.ExecCommandResponse} returns this
 */
proto.api_container_api.ExecCommandResponse.prototype.setLogOutput = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, ""),
    port: jspb.Message.getFieldWithDefault(msg, 2, 0),
    path: jspb.Message.getFieldWithDefault(msg, 3, ""),
    initialDelayMilliseconds: jspb.Message.getFieldWithDefault(msg, 4, 0),
    retries: jspb.Message.getFieldWithDefault(msg, 5, 0),
    retriesDelayMilliseconds: jspb.Message.getFieldWithDefault(msg, 6, 0),
    bodyText: jspb.Message.getFieldWithDefault(msg, 7, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs;
  return proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setPort(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setInitialDelayMilliseconds(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setRetries(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setRetriesDelayMilliseconds(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setBodyText(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPort();
  if (f !== 0) {
    writer.writeUint32(
      2,
      f
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getInitialDelayMilliseconds();
  if (f !== 0) {
    writer.writeUint32(
      4,
      f
    );
  }
  f = message.getRetries();
  if (f !== 0) {
    writer.writeUint32(
      5,
      f
    );
  }
  f = message.getRetriesDelayMilliseconds();
  if (f !== 0) {
    writer.writeUint32(
      6,
      f
    );
  }
  f = message.getBodyText();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional uint32 port = 2;
 * @return {number}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setPort = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string path = 3;
 * @return {string}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional uint32 initial_delay_milliseconds = 4;
 * @return {number}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getInitialDelayMilliseconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setInitialDelayMilliseconds = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional uint32 retries = 5;
 * @return {number}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getRetries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setRetries = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional uint32 retries_delay_milliseconds = 6;
 * @return {number}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getRetriesDelayMilliseconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setRetriesDelayMilliseconds = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional string body_text = 7;
 * @return {string}
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.getBodyText = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpGetEndpointAvailabilityArgs.prototype.setBodyText = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, ""),
    port: jspb.Message.getFieldWithDefault(msg, 2, 0),
    path: jspb.Message.getFieldWithDefault(msg, 3, ""),
    requestBody: jspb.Message.getFieldWithDefault(msg, 4, ""),
    initialDelayMilliseconds: jspb.Message.getFieldWithDefault(msg, 5, 0),
    retries: jspb.Message.getFieldWithDefault(msg, 6, 0),
    retriesDelayMilliseconds: jspb.Message.getFieldWithDefault(msg, 7, 0),
    bodyText: jspb.Message.getFieldWithDefault(msg, 8, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs;
  return proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setPort(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setRequestBody(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setInitialDelayMilliseconds(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setRetries(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setRetriesDelayMilliseconds(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setBodyText(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPort();
  if (f !== 0) {
    writer.writeUint32(
      2,
      f
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getRequestBody();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getInitialDelayMilliseconds();
  if (f !== 0) {
    writer.writeUint32(
      5,
      f
    );
  }
  f = message.getRetries();
  if (f !== 0) {
    writer.writeUint32(
      6,
      f
    );
  }
  f = message.getRetriesDelayMilliseconds();
  if (f !== 0) {
    writer.writeUint32(
      7,
      f
    );
  }
  f = message.getBodyText();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional uint32 port = 2;
 * @return {number}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setPort = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string path = 3;
 * @return {string}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string request_body = 4;
 * @return {string}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getRequestBody = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setRequestBody = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional uint32 initial_delay_milliseconds = 5;
 * @return {number}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getInitialDelayMilliseconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setInitialDelayMilliseconds = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional uint32 retries = 6;
 * @return {number}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getRetries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setRetries = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional uint32 retries_delay_milliseconds = 7;
 * @return {number}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getRetriesDelayMilliseconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setRetriesDelayMilliseconds = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional string body_text = 8;
 * @return {string}
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.getBodyText = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs} returns this
 */
proto.api_container_api.WaitForHttpPostEndpointAvailabilityArgs.prototype.setBodyText = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StreamedDataChunk.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StreamedDataChunk.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StreamedDataChunk} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StreamedDataChunk.toObject = function(includeInstance, msg) {
  var f, obj = {
    data: msg.getData_asB64(),
    previousChunkHash: jspb.Message.getFieldWithDefault(msg, 2, ""),
    metadata: (f = msg.getMetadata()) && proto.api_container_api.DataChunkMetadata.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StreamedDataChunk}
 */
proto.api_container_api.StreamedDataChunk.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StreamedDataChunk;
  return proto.api_container_api.StreamedDataChunk.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StreamedDataChunk} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StreamedDataChunk}
 */
proto.api_container_api.StreamedDataChunk.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setData(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPreviousChunkHash(value);
      break;
    case 3:
      var value = new proto.api_container_api.DataChunkMetadata;
      reader.readMessage(value,proto.api_container_api.DataChunkMetadata.deserializeBinaryFromReader);
      msg.setMetadata(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StreamedDataChunk.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StreamedDataChunk.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StreamedDataChunk} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StreamedDataChunk.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
  f = message.getPreviousChunkHash();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMetadata();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.api_container_api.DataChunkMetadata.serializeBinaryToWriter
    );
  }
};


/**
 * optional bytes data = 1;
 * @return {string}
 */
proto.api_container_api.StreamedDataChunk.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes data = 1;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.api_container_api.StreamedDataChunk.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.api_container_api.StreamedDataChunk.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.api_container_api.StreamedDataChunk} returns this
 */
proto.api_container_api.StreamedDataChunk.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
};


/**
 * optional string previous_chunk_hash = 2;
 * @return {string}
 */
proto.api_container_api.StreamedDataChunk.prototype.getPreviousChunkHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StreamedDataChunk} returns this
 */
proto.api_container_api.StreamedDataChunk.prototype.setPreviousChunkHash = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional DataChunkMetadata metadata = 3;
 * @return {?proto.api_container_api.DataChunkMetadata}
 */
proto.api_container_api.StreamedDataChunk.prototype.getMetadata = function() {
  return /** @type{?proto.api_container_api.DataChunkMetadata} */ (
    jspb.Message.getWrapperField(this, proto.api_container_api.DataChunkMetadata, 3));
};


/**
 * @param {?proto.api_container_api.DataChunkMetadata|undefined} value
 * @return {!proto.api_container_api.StreamedDataChunk} returns this
*/
proto.api_container_api.StreamedDataChunk.prototype.setMetadata = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.api_container_api.StreamedDataChunk} returns this
 */
proto.api_container_api.StreamedDataChunk.prototype.clearMetadata = function() {
  return this.setMetadata(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.api_container_api.StreamedDataChunk.prototype.hasMetadata = function() {
  return jspb.Message.getField(this, 3) != null;
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.DataChunkMetadata.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.DataChunkMetadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.DataChunkMetadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DataChunkMetadata.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.DataChunkMetadata}
 */
proto.api_container_api.DataChunkMetadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.DataChunkMetadata;
  return proto.api_container_api.DataChunkMetadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.DataChunkMetadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.DataChunkMetadata}
 */
proto.api_container_api.DataChunkMetadata.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.DataChunkMetadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.DataChunkMetadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.DataChunkMetadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DataChunkMetadata.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.api_container_api.DataChunkMetadata.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.DataChunkMetadata} returns this
 */
proto.api_container_api.DataChunkMetadata.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.UploadFilesArtifactArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.UploadFilesArtifactArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UploadFilesArtifactArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    data: msg.getData_asB64(),
    name: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.UploadFilesArtifactArgs}
 */
proto.api_container_api.UploadFilesArtifactArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.UploadFilesArtifactArgs;
  return proto.api_container_api.UploadFilesArtifactArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.UploadFilesArtifactArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.UploadFilesArtifactArgs}
 */
proto.api_container_api.UploadFilesArtifactArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setData(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.UploadFilesArtifactArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.UploadFilesArtifactArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UploadFilesArtifactArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional bytes data = 1;
 * @return {string}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes data = 1;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.api_container_api.UploadFilesArtifactArgs} returns this
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.UploadFilesArtifactArgs} returns this
 */
proto.api_container_api.UploadFilesArtifactArgs.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.UploadFilesArtifactResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.UploadFilesArtifactResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UploadFilesArtifactResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    uuid: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.UploadFilesArtifactResponse}
 */
proto.api_container_api.UploadFilesArtifactResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.UploadFilesArtifactResponse;
  return proto.api_container_api.UploadFilesArtifactResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.UploadFilesArtifactResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.UploadFilesArtifactResponse}
 */
proto.api_container_api.UploadFilesArtifactResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUuid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.UploadFilesArtifactResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.UploadFilesArtifactResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.UploadFilesArtifactResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string uuid = 1;
 * @return {string}
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.getUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.UploadFilesArtifactResponse} returns this
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.setUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.UploadFilesArtifactResponse} returns this
 */
proto.api_container_api.UploadFilesArtifactResponse.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.DownloadFilesArtifactArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.DownloadFilesArtifactArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DownloadFilesArtifactArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    identifier: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.DownloadFilesArtifactArgs}
 */
proto.api_container_api.DownloadFilesArtifactArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.DownloadFilesArtifactArgs;
  return proto.api_container_api.DownloadFilesArtifactArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.DownloadFilesArtifactArgs}
 */
proto.api_container_api.DownloadFilesArtifactArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setIdentifier(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.DownloadFilesArtifactArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.DownloadFilesArtifactArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.DownloadFilesArtifactArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DownloadFilesArtifactArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string identifier = 1;
 * @return {string}
 */
proto.api_container_api.DownloadFilesArtifactArgs.prototype.getIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.DownloadFilesArtifactArgs} returns this
 */
proto.api_container_api.DownloadFilesArtifactArgs.prototype.setIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.DownloadFilesArtifactResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.DownloadFilesArtifactResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DownloadFilesArtifactResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    data: msg.getData_asB64()
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.DownloadFilesArtifactResponse}
 */
proto.api_container_api.DownloadFilesArtifactResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.DownloadFilesArtifactResponse;
  return proto.api_container_api.DownloadFilesArtifactResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.DownloadFilesArtifactResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.DownloadFilesArtifactResponse}
 */
proto.api_container_api.DownloadFilesArtifactResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setData(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.DownloadFilesArtifactResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.DownloadFilesArtifactResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.DownloadFilesArtifactResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
};


/**
 * optional bytes data = 1;
 * @return {string}
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes data = 1;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.api_container_api.DownloadFilesArtifactResponse} returns this
 */
proto.api_container_api.DownloadFilesArtifactResponse.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StoreWebFilesArtifactArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreWebFilesArtifactArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    url: jspb.Message.getFieldWithDefault(msg, 1, ""),
    name: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StoreWebFilesArtifactArgs}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StoreWebFilesArtifactArgs;
  return proto.api_container_api.StoreWebFilesArtifactArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StoreWebFilesArtifactArgs}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUrl(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StoreWebFilesArtifactArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StoreWebFilesArtifactArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreWebFilesArtifactArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUrl();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string url = 1;
 * @return {string}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreWebFilesArtifactArgs} returns this
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.setUrl = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreWebFilesArtifactArgs} returns this
 */
proto.api_container_api.StoreWebFilesArtifactArgs.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StoreWebFilesArtifactResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StoreWebFilesArtifactResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StoreWebFilesArtifactResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreWebFilesArtifactResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    uuid: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StoreWebFilesArtifactResponse}
 */
proto.api_container_api.StoreWebFilesArtifactResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StoreWebFilesArtifactResponse;
  return proto.api_container_api.StoreWebFilesArtifactResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StoreWebFilesArtifactResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StoreWebFilesArtifactResponse}
 */
proto.api_container_api.StoreWebFilesArtifactResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StoreWebFilesArtifactResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StoreWebFilesArtifactResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StoreWebFilesArtifactResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreWebFilesArtifactResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string uuid = 1;
 * @return {string}
 */
proto.api_container_api.StoreWebFilesArtifactResponse.prototype.getUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreWebFilesArtifactResponse} returns this
 */
proto.api_container_api.StoreWebFilesArtifactResponse.prototype.setUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StoreFilesArtifactFromServiceArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    serviceIdentifier: jspb.Message.getFieldWithDefault(msg, 1, ""),
    sourcePath: jspb.Message.getFieldWithDefault(msg, 2, ""),
    name: jspb.Message.getFieldWithDefault(msg, 3, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceArgs}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StoreFilesArtifactFromServiceArgs;
  return proto.api_container_api.StoreFilesArtifactFromServiceArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceArgs}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setServiceIdentifier(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSourcePath(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StoreFilesArtifactFromServiceArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getServiceIdentifier();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSourcePath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string service_identifier = 1;
 * @return {string}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.getServiceIdentifier = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} returns this
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.setServiceIdentifier = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string source_path = 2;
 * @return {string}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.getSourcePath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} returns this
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.setSourcePath = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string name = 3;
 * @return {string}
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceArgs} returns this
 */
proto.api_container_api.StoreFilesArtifactFromServiceArgs.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.StoreFilesArtifactFromServiceResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    uuid: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceResponse}
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.StoreFilesArtifactFromServiceResponse;
  return proto.api_container_api.StoreFilesArtifactFromServiceResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceResponse}
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.StoreFilesArtifactFromServiceResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.StoreFilesArtifactFromServiceResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string uuid = 1;
 * @return {string}
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.prototype.getUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.StoreFilesArtifactFromServiceResponse} returns this
 */
proto.api_container_api.StoreFilesArtifactFromServiceResponse.prototype.setUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RenderTemplatesToFilesArtifactArgs.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.toObject = function(includeInstance, msg) {
  var f, obj = {
    templatesAndDataByDestinationRelFilepathMap: (f = msg.getTemplatesAndDataByDestinationRelFilepathMap()) ? f.toObject(includeInstance, proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.toObject) : [],
    name: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RenderTemplatesToFilesArtifactArgs;
  return proto.api_container_api.RenderTemplatesToFilesArtifactArgs.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getTemplatesAndDataByDestinationRelFilepathMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.deserializeBinaryFromReader, "", new proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData());
         });
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RenderTemplatesToFilesArtifactArgs.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTemplatesAndDataByDestinationRelFilepathMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.serializeBinaryToWriter);
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.toObject = function(includeInstance, msg) {
  var f, obj = {
    template: jspb.Message.getFieldWithDefault(msg, 1, ""),
    dataAsJson: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData;
  return proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setTemplate(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDataAsJson(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTemplate();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getDataAsJson();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string template = 1;
 * @return {string}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.getTemplate = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData} returns this
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.setTemplate = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string data_as_json = 2;
 * @return {string}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.getDataAsJson = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData} returns this
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData.prototype.setDataAsJson = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * map<string, TemplateAndData> templates_and_data_by_destination_rel_filepath = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData>}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.getTemplatesAndDataByDestinationRelFilepathMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.api_container_api.RenderTemplatesToFilesArtifactArgs.TemplateAndData));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} returns this
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.clearTemplatesAndDataByDestinationRelFilepathMap = function() {
  this.getTemplatesAndDataByDestinationRelFilepathMap().clear();
  return this;};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactArgs} returns this
 */
proto.api_container_api.RenderTemplatesToFilesArtifactArgs.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.RenderTemplatesToFilesArtifactResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    uuid: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.RenderTemplatesToFilesArtifactResponse;
  return proto.api_container_api.RenderTemplatesToFilesArtifactResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setUuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.RenderTemplatesToFilesArtifactResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUuid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string uuid = 1;
 * @return {string}
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.prototype.getUuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.RenderTemplatesToFilesArtifactResponse} returns this
 */
proto.api_container_api.RenderTemplatesToFilesArtifactResponse.prototype.setUuid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.FilesArtifactNameAndUuid.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.FilesArtifactNameAndUuid} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.FilesArtifactNameAndUuid.toObject = function(includeInstance, msg) {
  var f, obj = {
    filename: jspb.Message.getFieldWithDefault(msg, 1, ""),
    fileuuid: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.FilesArtifactNameAndUuid}
 */
proto.api_container_api.FilesArtifactNameAndUuid.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.FilesArtifactNameAndUuid;
  return proto.api_container_api.FilesArtifactNameAndUuid.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.FilesArtifactNameAndUuid} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.FilesArtifactNameAndUuid}
 */
proto.api_container_api.FilesArtifactNameAndUuid.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilename(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setFileuuid(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.FilesArtifactNameAndUuid.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.FilesArtifactNameAndUuid} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.FilesArtifactNameAndUuid.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFilename();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getFileuuid();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string fileName = 1;
 * @return {string}
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.getFilename = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.FilesArtifactNameAndUuid} returns this
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.setFilename = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string fileUuid = 2;
 * @return {string}
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.getFileuuid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.api_container_api.FilesArtifactNameAndUuid} returns this
 */
proto.api_container_api.FilesArtifactNameAndUuid.prototype.setFileuuid = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    fileNamesAndUuidsList: jspb.Message.toObjectList(msg.getFileNamesAndUuidsList(),
    proto.api_container_api.FilesArtifactNameAndUuid.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse;
  return proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.api_container_api.FilesArtifactNameAndUuid;
      reader.readMessage(value,proto.api_container_api.FilesArtifactNameAndUuid.deserializeBinaryFromReader);
      msg.addFileNamesAndUuids(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFileNamesAndUuidsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.api_container_api.FilesArtifactNameAndUuid.serializeBinaryToWriter
    );
  }
};


/**
 * repeated FilesArtifactNameAndUuid file_names_and_uuids = 1;
 * @return {!Array<!proto.api_container_api.FilesArtifactNameAndUuid>}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.getFileNamesAndUuidsList = function() {
  return /** @type{!Array<!proto.api_container_api.FilesArtifactNameAndUuid>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.api_container_api.FilesArtifactNameAndUuid, 1));
};


/**
 * @param {!Array<!proto.api_container_api.FilesArtifactNameAndUuid>} value
 * @return {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse} returns this
*/
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.setFileNamesAndUuidsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.api_container_api.FilesArtifactNameAndUuid=} opt_value
 * @param {number=} opt_index
 * @return {!proto.api_container_api.FilesArtifactNameAndUuid}
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.addFileNamesAndUuids = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.api_container_api.FilesArtifactNameAndUuid, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse} returns this
 */
proto.api_container_api.ListFilesArtifactNamesAndUuidsResponse.prototype.clearFileNamesAndUuidsList = function() {
  return this.setFileNamesAndUuidsList([]);
};


goog.object.extend(exports, proto.api_container_api);
