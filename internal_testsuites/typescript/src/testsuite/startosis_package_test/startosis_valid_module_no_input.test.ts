import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    EMPTY_RUN_PARAMS,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";
import {err} from "neverthrow";

const VALID_PACKAGE_NO_PACKAGE_INPUT_TEST_NAME = "valid-package-no-input"
const VALID_PACKAGE_NO_PACKAGE_INPUT_REL_PATH = "../../../../starlark/valid-kurtosis-package-no-input"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid Starlark package with input", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_PACKAGE_NO_PACKAGE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, VALID_PACKAGE_NO_PACKAGE_INPUT_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const outputStream = await enclaveContext.runStarlarkPackage(packageRootPath, EMPTY_RUN_PARAMS, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            throw err(new Error(`An error occurred execute Starlark package '${packageRootPath}'`));
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

test("Test valid Starlark package with input - passing params also works", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_PACKAGE_NO_PACKAGE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, VALID_PACKAGE_NO_PACKAGE_INPUT_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const params = `{"greetings": "bonjour!"}`
        const outputStream = await enclaveContext.runStarlarkPackage(packageRootPath, params, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            throw err(new Error(`An error occurred execute Starlark package '${packageRootPath}'`));
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
