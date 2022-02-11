import * as jspb from 'google-protobuf'

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class ExecuteArgs extends jspb.Message {
  getParamsJson(): string;
  setParamsJson(value: string): ExecuteArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteArgs): ExecuteArgs.AsObject;
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
  setResponseJson(value: string): ExecuteResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteResponse): ExecuteResponse.AsObject;
  static serializeBinaryToWriter(message: ExecuteResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteResponse;
  static deserializeBinaryFromReader(message: ExecuteResponse, reader: jspb.BinaryReader): ExecuteResponse;
}

export namespace ExecuteResponse {
  export type AsObject = {
    responseJson: string,
  }
}

