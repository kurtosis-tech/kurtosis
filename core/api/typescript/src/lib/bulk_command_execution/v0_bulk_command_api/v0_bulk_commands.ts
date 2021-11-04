import { ExecCommandArgs, ExecuteBulkCommandsArgs, ExecuteModuleArgs, LoadModuleArgs, RegisterServiceArgs, RemoveServiceArgs, RepartitionArgs, StartServiceArgs, WaitForHttpGetEndpointAvailabilityArgs, WaitForHttpPostEndpointAvailabilityArgs } from "../../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { V0CommandTypeVisitor, V0CommandType } from "./v0_command_types";
import { ok, err, Result } from "neverthrow";
import * as protobuf from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

// ====================================================================================================
//                                   Command Arg Deserialization Visitor
// ====================================================================================================

// Visitor that will be used to deserialize command args into
class CmdArgDeserializingVisitor implements V0CommandTypeVisitor {
    private static safeJsonParse = Result.fromThrowable(JSON.parse, CmdArgDeserializingVisitor.parseUnknownExceptionValueToError);

    private readonly bytesToDeserialize: string;
    private deserializedCommandArgsPtr?: protobuf.Message;

    constructor (bytesToDeserialize: string) {
        this.bytesToDeserialize = bytesToDeserialize;
    }

    public visitLoadModule(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: LoadModuleArgs = Object.assign(new LoadModuleArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitExecuteModule(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: ExecuteModuleArgs = Object.assign(new ExecuteModuleArgs(), deserializationResult.value);

    	this.deserializedCommandArgsPtr = args;
    	return ok(null);
    }

    public visitRegisterService(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: RegisterServiceArgs = Object.assign(new RegisterServiceArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitStartService(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: StartServiceArgs = Object.assign(new StartServiceArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitRemoveService(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: RemoveServiceArgs = Object.assign(new RemoveServiceArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitRepartition(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: RepartitionArgs = Object.assign(new RepartitionArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitExecCommand(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: ExecCommandArgs = Object.assign(new ExecCommandArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitWaitForHttpGetEndpointAvailability(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: WaitForHttpGetEndpointAvailabilityArgs = Object.assign(new WaitForHttpGetEndpointAvailabilityArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitWaitForHttpPostEndpointAvailability(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: WaitForHttpPostEndpointAvailabilityArgs = Object.assign(new WaitForHttpPostEndpointAvailabilityArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitExecuteBulkCommands(): Result<null, Error> {
        const deserializationResult: Result<any, Error> = CmdArgDeserializingVisitor.safeJsonParse(this.bytesToDeserialize);
        if (deserializationResult.isErr()) {
            return err(deserializationResult.error);
        }
        const args: ExecuteBulkCommandsArgs = Object.assign(new ExecuteBulkCommandsArgs(), deserializationResult.value);

        this.deserializedCommandArgsPtr = args;
        return ok(null);
    }

    public visitGetServices(): Result<null, Error> {
        this.deserializedCommandArgsPtr = new google_protobuf_empty_pb.Empty()
        return ok(null);
    }

    public visitGetModules(): Result<null, Error> {
        this.deserializedCommandArgsPtr = new google_protobuf_empty_pb.Empty()
        return ok(null);
    }

    public getDeserializedCommandArgs(): Result<protobuf.Message, Error> {
        if (!this.deserializedCommandArgsPtr) {
            return err(new Error("Deserialized command args pointer was falsy; this indicates that it was never set through a visitor method, which should never happen"));
        }
        return ok(this.deserializedCommandArgsPtr!);
    }

    private static parseUnknownExceptionValueToError(value: unknown): Error {
        if (value instanceof Error) {
            return value;
        }
        return new Error("Received an unknown exception value that wasn't an error: " + value);
    }
}

// ====================================================================================================
//                                        Serializable Command
// ====================================================================================================

// Used for serializing
export class V0SerializableCommand {
    private readonly type: V0CommandType;
    
    // The only allowed objects here are from the bindings generated from the .proto file
    private readonly argsPtr: protobuf.Message;

    constructor (type: V0CommandType, argsPtr: protobuf.Message) {
        this.type = type;
        this.argsPtr = argsPtr;
    } 
       
    public getType(): V0CommandType {
        return this.type;
    }

    public getArgsPtr(): protobuf.Message {
        return this.argsPtr;
    }

}

// // ====================================================================================================
// //                                   Bulk Commands Package
// // ====================================================================================================

export class V0BulkCommands {
    private readonly commands: V0SerializableCommand[];

    constructor (commands: V0SerializableCommand[]) {
        this.commands = commands;
    }
}
