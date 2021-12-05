// package: module_api
// file: executable_module_service.proto

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

export class ExecuteArgs extends jspb.Message {
  getParamsJson(): string;
  setParamsJson(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteArgs): ExecuteArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecuteArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteArgs;
  static deserializeBinaryFromReader(message: ExecuteArgs, reader: jspb.BinaryReader): ExecuteArgs;
}

export namespace ExecuteArgs {
  export type AsObject = {
    paramsJson: string,
  }
}

export class ExecuteResponse extends jspb.Message {
  getResponseJson(): string;
  setResponseJson(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteResponse): ExecuteResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecuteResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteResponse;
  static deserializeBinaryFromReader(message: ExecuteResponse, reader: jspb.BinaryReader): ExecuteResponse;
}

export namespace ExecuteResponse {
  export type AsObject = {
    responseJson: string,
  }
}

