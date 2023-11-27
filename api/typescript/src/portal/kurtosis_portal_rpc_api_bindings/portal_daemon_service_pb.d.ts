import * as jspb from 'google-protobuf'



export class ForwardUserServicePortArgs extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): ForwardUserServicePortArgs;

  getServiceId(): string;
  setServiceId(value: string): ForwardUserServicePortArgs;

  getPortId(): string;
  setPortId(value: string): ForwardUserServicePortArgs;

  getLocalPortNumber(): number;
  setLocalPortNumber(value: number): ForwardUserServicePortArgs;
  hasLocalPortNumber(): boolean;
  clearLocalPortNumber(): ForwardUserServicePortArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ForwardUserServicePortArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ForwardUserServicePortArgs): ForwardUserServicePortArgs.AsObject;
  static serializeBinaryToWriter(message: ForwardUserServicePortArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ForwardUserServicePortArgs;
  static deserializeBinaryFromReader(message: ForwardUserServicePortArgs, reader: jspb.BinaryReader): ForwardUserServicePortArgs;
}

export namespace ForwardUserServicePortArgs {
  export type AsObject = {
    enclaveId: string,
    serviceId: string,
    portId: string,
    localPortNumber?: number,
  }

  export enum LocalPortNumberCase { 
    _LOCAL_PORT_NUMBER_NOT_SET = 0,
    LOCAL_PORT_NUMBER = 4,
  }
}

export class ForwardPortResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ForwardPortResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ForwardPortResponse): ForwardPortResponse.AsObject;
  static serializeBinaryToWriter(message: ForwardPortResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ForwardPortResponse;
  static deserializeBinaryFromReader(message: ForwardPortResponse, reader: jspb.BinaryReader): ForwardPortResponse;
}

export namespace ForwardPortResponse {
  export type AsObject = {
  }
}

export class ForwardUserServicePortResponse extends jspb.Message {
  getLocalPortNumber(): number;
  setLocalPortNumber(value: number): ForwardUserServicePortResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ForwardUserServicePortResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ForwardUserServicePortResponse): ForwardUserServicePortResponse.AsObject;
  static serializeBinaryToWriter(message: ForwardUserServicePortResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ForwardUserServicePortResponse;
  static deserializeBinaryFromReader(message: ForwardUserServicePortResponse, reader: jspb.BinaryReader): ForwardUserServicePortResponse;
}

export namespace ForwardUserServicePortResponse {
  export type AsObject = {
    localPortNumber: number,
  }
}

export class PortalPing extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PortalPing.AsObject;
  static toObject(includeInstance: boolean, msg: PortalPing): PortalPing.AsObject;
  static serializeBinaryToWriter(message: PortalPing, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PortalPing;
  static deserializeBinaryFromReader(message: PortalPing, reader: jspb.BinaryReader): PortalPing;
}

export namespace PortalPing {
  export type AsObject = {
  }
}

export class PortalPong extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PortalPong.AsObject;
  static toObject(includeInstance: boolean, msg: PortalPong): PortalPong.AsObject;
  static serializeBinaryToWriter(message: PortalPong, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PortalPong;
  static deserializeBinaryFromReader(message: PortalPong, reader: jspb.BinaryReader): PortalPong;
}

export namespace PortalPong {
  export type AsObject = {
  }
}

