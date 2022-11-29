import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    EMPTY_EXECUTE_PARAMS,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";
import {err} from "neverthrow";

const VALID_MODULE_NO_MODULE_INPUT_TEST_NAME = "valid-module-no-input"
const VALID_MODULE_NO_MODULE_INPUT_REL_PATH = "../../../../startosis/valid-kurtosis-module-no-input"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid kurtosis module with input", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const outputStream = await enclaveContext.runStarlarkPackage(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            throw err(new Error(`An error occurred execute startosis module '${moduleRootPath}'`));
        }
        const [scriptOutput, _, interpretationError, validationErrors, executionError] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptOutput = "Hello world!\n"

        expect(scriptOutput).toEqual(expectedScriptOutput)

        expect(interpretationError).toBeUndefined()
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()
    } finally {
        stopEnclaveFunction()
    }
})

test("Test valid kurtosis module with input - passing params also works", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const params = `{"greetings": "bonjour!"}`
        const outputStream = await enclaveContext.runStarlarkPackage(moduleRootPath, params, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            throw err(new Error(`An error occurred execute startosis module '${moduleRootPath}'`));
        }
        const [scriptOutput, _, interpretationError, validationErrors, executionError] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptOutput = "Hello world!\n"

        expect(scriptOutput).toEqual(expectedScriptOutput)

        expect(interpretationError).toBeUndefined()
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()
    } finally {
        stopEnclaveFunction()
    }
})
