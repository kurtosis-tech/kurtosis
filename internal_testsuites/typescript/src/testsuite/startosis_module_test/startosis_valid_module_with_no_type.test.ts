import log from "loglevel";

import {createEnclave} from "../../test_helpers/enclave_setup";
import * as path from "path";
import {DEFAULT_DRY_RUN, EMPTY_EXECUTE_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import {generateScriptOutput, readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

const VALID_MODULE_NO_TYPE_TEST_NAME = "valid-module-no-type";
const VALID_MODULE_NO_TYPE_REL_PATH = "../../../../startosis/valid-kurtosis-module-no-type"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with no type", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const outputStream = await enclaveContext.executeKurtosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if (outputStream.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw outputStream.error
        }
        const [interpretationError, validationErrors, executionError, instructions] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptOutput = "Hello World!\n"

        expect(generateScriptOutput(instructions)).toEqual(expectedScriptOutput)

        expect(interpretationError).toBeUndefined()
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()
    } finally {
        stopEnclaveFunction()
    }
})
