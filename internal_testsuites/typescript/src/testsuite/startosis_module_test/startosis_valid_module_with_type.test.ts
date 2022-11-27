import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {generateScriptOutput, readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

const VALID_MODULE_WITH_TYPES_TEST_NAME = "valid-module-with-types";
const VALID_MODULE_WITH_TYPES_REL_PATH = "../../../../startosis/valid-kurtosis-module-with-types"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with types", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_WITH_TYPES_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_WITH_TYPES_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const outputStream = await enclaveContext.executeKurtosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if (outputStream.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw outputStream.error
        }
        const [interpretationError, validationErrors, executionError, instructions] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptOutput = "Bonjour!\nHello World!\n"

        expect(generateScriptOutput(instructions)).toEqual(expectedScriptOutput)

        expect(interpretationError).toBeUndefined()
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()
    } finally {
        stopEnclaveFunction()
    }
})
