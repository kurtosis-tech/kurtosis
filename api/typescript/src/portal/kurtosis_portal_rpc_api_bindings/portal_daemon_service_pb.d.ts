import * as jspb from 'google-protobuf'



export class EnclaveServicePortId extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): EnclaveServicePortId;

  getServiceId(): string;
  setServiceId(value: string): EnclaveServicePortId;
  hasServiceId(): boolean;
  clearServiceId(): EnclaveServicePortId;

  getPortId(): string;
  setPortId(value: string): EnclaveServicePortId;
  hasPortId(): boolean;
  clearPortId(): EnclaveServicePortId;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveServicePortId.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveServicePortId): EnclaveServicePortId.AsObject;
  static serializeBinaryToWriter(message: EnclaveServicePortId, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveServicePortId;
  static deserializeBinaryFromReader(message: EnclaveServicePortId, reader: jspb.BinaryReader): EnclaveServicePortId;
}

export namespace EnclaveServicePortId {
  export type AsObject = {
    enclaveId: string,
    serviceId?: string,
    portId?: string,
  }

  export enum ServiceIdCase { 
    _SERVICE_ID_NOT_SET = 0,
    SERVICE_ID = 2,
  }

  export enum PortIdCase { 
    _PORT_ID_NOT_SET = 0,
    PORT_ID = 3,
  }
}

export class ForwardedServicePort extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): ForwardedServicePort;

  getServiceId(): string;
  setServiceId(value: string): ForwardedServicePort;

  getPortId(): string;
  setPortId(value: string): ForwardedServicePort;

  getLocalPortNumber(): number;
  setLocalPortNumber(value: number): ForwardedServicePort;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ForwardedServicePort.AsObject;
  static toObject(includeInstance: boolean, msg: ForwardedServicePort): ForwardedServicePort.AsObject;
  static serializeBinaryToWriter(message: ForwardedServicePort, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ForwardedServicePort;
  static deserializeBinaryFromReader(message: ForwardedServicePort, reader: jspb.BinaryReader): ForwardedServicePort;
}

export namespace ForwardedServicePort {
  export type AsObject = {
    enclaveId: string,
    serviceId: string,
    portId: string,
    localPortNumber: number,
  }
}

export class CreateUserServicePortForwardArgs extends jspb.Message {
  getEnclaveserviceportid(): EnclaveServicePortId | undefined;
  setEnclaveserviceportid(value?: EnclaveServicePortId): CreateUserServicePortForwardArgs;
  hasEnclaveserviceportid(): boolean;
  clearEnclaveserviceportid(): CreateUserServicePortForwardArgs;

  getLocalPortNumber(): number;
  setLocalPortNumber(value: number): CreateUserServicePortForwardArgs;
  hasLocalPortNumber(): boolean;
  clearLocalPortNumber(): CreateUserServicePortForwardArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateUserServicePortForwardArgs.AsObject;
  static toObject(includeInstance: boolean, msg: CreateUserServicePortForwardArgs): CreateUserServicePortForwardArgs.AsObject;
  static serializeBinaryToWriter(message: CreateUserServicePortForwardArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateUserServicePortForwardArgs;
  static deserializeBinaryFromReader(message: CreateUserServicePortForwardArgs, reader: jspb.BinaryReader): CreateUserServicePortForwardArgs;
}

export namespace CreateUserServicePortForwardArgs {
  export type AsObject = {
    enclaveserviceportid?: EnclaveServicePortId.AsObject,
    localPortNumber?: number,
  }

  export enum LocalPortNumberCase { 
    _LOCAL_PORT_NUMBER_NOT_SET = 0,
    LOCAL_PORT_NUMBER = 2,
  }
}

export class CreateUserServicePortForwardResponse extends jspb.Message {
  getForwardedPortNumbersList(): Array<ForwardedServicePort>;
  setForwardedPortNumbersList(value: Array<ForwardedServicePort>): CreateUserServicePortForwardResponse;
  clearForwardedPortNumbersList(): CreateUserServicePortForwardResponse;
  addForwardedPortNumbers(value?: ForwardedServicePort, index?: number): ForwardedServicePort;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateUserServicePortForwardResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateUserServicePortForwardResponse): CreateUserServicePortForwardResponse.AsObject;
  static serializeBinaryToWriter(message: CreateUserServicePortForwardResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateUserServicePortForwardResponse;
  static deserializeBinaryFromReader(message: CreateUserServicePortForwardResponse, reader: jspb.BinaryReader): CreateUserServicePortForwardResponse;
}

export namespace CreateUserServicePortForwardResponse {
  export type AsObject = {
    forwardedPortNumbersList: Array<ForwardedServicePort.AsObject>,
  }
}

export class RemoveUserServicePortForwardResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveUserServicePortForwardResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveUserServicePortForwardResponse): RemoveUserServicePortForwardResponse.AsObject;
  static serializeBinaryToWriter(message: RemoveUserServicePortForwardResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveUserServicePortForwardResponse;
  static deserializeBinaryFromReader(message: RemoveUserServicePortForwardResponse, reader: jspb.BinaryReader): RemoveUserServicePortForwardResponse;
}

export namespace RemoveUserServicePortForwardResponse {
  export type AsObject = {
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

