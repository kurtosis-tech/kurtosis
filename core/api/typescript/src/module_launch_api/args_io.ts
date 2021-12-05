
// All module containers accept exactly one environment variable, which contains the serialized params that

import log = require("loglevel");
import { err, ok, Result } from "neverthrow";
import { ModuleContainerArgs } from "./module_container_args";

// dictate how the module container ought to behave
const SERIALIZED_ARGS_ENV_VAR = "SERIALIZED_ARGS";
const JAVASCRIPT_STRING_VAR_TYPE = "string";

const safeJsonParse = Result.fromThrowable(JSON.parse, (value: unknown): Error => {
    if (value instanceof Error) {
        return value;
    }
    return new Error("Received an unknown exception value that wasn't an error: " + value);
});

// Intended to be used when starting the container - gets the environment variables that the container should be started with
export function getEnvFromArgs(args: ModuleContainerArgs): Result<Map<string, string>, Error> {
    let argsStr: string;
    try {
        argsStr = JSON.stringify(args);
    } catch (jsonErr: any) {
        // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
        // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
        if (jsonErr && jsonErr.stack && jsonErr.message) {
            return err(jsonErr as Error);
        }
        return err(new Error("Stringify-ing ModuleContainerArgs object threw an exception, but " +
            "it's not an Error so we can't report any more information than this"));
    }
    const result = new Map<string, string>();
    result.set(SERIALIZED_ARGS_ENV_VAR, argsStr);
    return ok(result);
}

export function getArgsFromEnv(): Result<ModuleContainerArgs, Error> {
    const serializedParamsStr = process.env[SERIALIZED_ARGS_ENV_VAR];
    if (serializedParamsStr === undefined) {
		return err(new Error(`No serialized args variable '${SERIALIZED_ARGS_ENV_VAR}' defined`))
    }
    if (serializedParamsStr === "") {
		return err(new Error(`Found serialized args environment variable '${SERIALIZED_ARGS_ENV_VAR}', but the value was empty`))
    }
    const deserializationResult = safeJsonParse(serializedParamsStr);
    if (deserializationResult.isErr()) {
        // TODO Replace with a stacktrace.Propagate equivalent
        log.error(`An error occurred deserializing the args JSON '${serializedParamsStr}'`);
        return err(deserializationResult.error);
    }
    const args = deserializationResult.value as ModuleContainerArgs;

	// Generic validation based on field type
    for (let [key, value] of Object.entries(args)) {
        // Ensure no empty strings
        if (typeof value === JAVASCRIPT_STRING_VAR_TYPE && (value as string).trim() === "") {
            return err(new Error(`JSON field '${key}' is whitespace or empty string`));
        }
    }

    return ok(args);
}
