import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {generateScriptOutput, readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with no module input type in types file - failure called with params", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const outputStream = await enclaveContext.executeKurtosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if (outputStream.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw outputStream.error
        }
        const [interpretationError, validationErrors, executionError, instructions] = await readStreamContentUntilClosed(outputStream.value);

        expect(interpretationError).not.toBeUndefined()
        expect(interpretationError?.getErrorMessage())
            .toContain("A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but 'ModuleInput' type is not defined in the module's 'types.proto' file.")

        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()

        expect(generateScriptOutput(instructions)).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
