import { ok, err, Result } from "neverthrow";

// We provide a visitor interface here so that:
//  1) all enum cases can be exhaustively handled
//  2) any changes in the enum will result in a compile break
export interface V0CommandTypeVisitor {
	visitLoadModule: () => Result<null, Error>;
	visitExecuteModule: () => Result<null, Error>;
	visitRegisterService: () => Result<null, Error>;
	visitStartService: () => Result<null, Error>;
	visitRemoveService: () => Result<null, Error>;
	visitRepartition: () => Result<null, Error>;
	visitExecCommand: () => Result<null, Error>;
	visitWaitForHttpGetEndpointAvailability: () => Result<null, Error>;
	visitWaitForHttpPostEndpointAvailability: () => Result<null, Error>;
	visitExecuteBulkCommands: () => Result<null, Error>;
	visitGetServices: () => Result<null, Error>;
	visitGetModules: () => Result<null, Error>;
}

// vvvvvvvvvvvvvvvvvvvv Update the visitor whenever you add an enum value!!! vvvvvvvvvvvvvvvvvvvvvvvvvvv
export enum V0CommandType {
	LoadModule = "LOAD_MODULE",
	ExecuteModule = "EXECUTE_MODULE",
	RegisterService = "REGISTER_SERVICE",
	StartService = "START_SERVICE",
	RemoveService = "REMOVE_SERVICE",
	Repartition = "REPARTITION",
	ExecCommand = "EXEC_COMMAND",
	WaitForHttpGetEndpointAvailability = "WAIT_FOR_HTTP_GET_ENDPOINT_AVAILABILITY",
	WaitForHttpPostEndpointAvailability = "WAIT_FOR_HTTP_POST_ENDPOINT_AVAILABILITY",
	ExecuteBulkCommands = "EXECUTE_BULK_COMMANDS",
	GetServices = "GET_SERVICES",
	GetModules = "GET_MODULES"
}
// ^^^^^^^^^^^^^^^^^^^^ Update the visitor whenever you add an enum value!!! ^^^^^^^^^^^^^^^^^^^^^^^^^^^

export namespace V0CommandType {
	export function acceptVisitor(commandType: V0CommandType, visitor: V0CommandTypeVisitor): Result<null, Error> {
		let result: Result<null, Error>;
		switch (commandType) {
			case V0CommandType.LoadModule:
				result = visitor.visitLoadModule();
				break;
			case V0CommandType.ExecuteModule:
				result = visitor.visitExecuteModule();
				break;
			case V0CommandType.RegisterService:
				result = visitor.visitRegisterService();
				break;
			case V0CommandType.StartService:
				result = visitor.visitStartService();
				break;
			case V0CommandType.RemoveService:
				result = visitor.visitRemoveService();
				break;
			case V0CommandType.Repartition:
				result = visitor.visitRepartition();
				break;
			case V0CommandType.ExecCommand:
				result = visitor.visitExecCommand();
				break;
			case V0CommandType.WaitForHttpGetEndpointAvailability:
				result = visitor.visitWaitForHttpGetEndpointAvailability();
				break;
			case V0CommandType.WaitForHttpPostEndpointAvailability:
				result = visitor.visitWaitForHttpPostEndpointAvailability();
				break;
			case V0CommandType.ExecuteBulkCommands:
				result = visitor.visitExecuteBulkCommands();
				break;
			case V0CommandType.GetServices:
				result = visitor.visitGetServices();
				break;
			case V0CommandType.GetModules:
				result = visitor.visitGetModules();
				break;
			default:
				return err(new Error("Unrecognized command type " + commandType));
		}
		if (!result.isOk()) {
			return err(result.error);
		}
		return ok(null);
	}
}
