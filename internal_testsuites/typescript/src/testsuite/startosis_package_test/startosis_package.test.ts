import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";

const VALID_PACKAGE_WITH_PACKAGE_INPUT_TEST_NAME = "valid-package-with-input"
const VALID_PACKAGE_WITH_PACKAGE_INPUT_REL_PATH = "../../../../starlark/valid-kurtosis-package-with-input"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid Starlark package with input", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_PACKAGE_WITH_PACKAGE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, VALID_PACKAGE_WITH_PACKAGE_INPUT_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const params = `{"greetings": "bonjour!"}`
        const runResult = await enclaveContext.runStarlarkPackageBlocking(packageRootPath, params, DEFAULT_DRY_RUN)

        if (runResult.isErr()) {
            log.error(`An error occurred execute Starlark package '${packageRootPath}'`);
            throw runResult.error
        }

        expect(runResult.value.interpretationError).toBeUndefined()
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()

        const expectedScriptOutput = "bonjour!\nHello World!\n{\n\t\"message\": \"Hello World!\"\n}\n"
        expect(runResult.value.runOutput).toEqual(expectedScriptOutput)
        expect(runResult.value.instructions).toHaveLength(2)
    } finally {
        stopEnclaveFunction()
    }
})

test("Test valid Starlark package with input - missing key in params", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_PACKAGE_WITH_PACKAGE_INPUT_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, VALID_PACKAGE_WITH_PACKAGE_INPUT_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const params = `{"hello": "world"}` // expecting key 'greetings' here
        const runResult = await enclaveContext.runStarlarkPackageBlocking(packageRootPath, params, DEFAULT_DRY_RUN)

        if (runResult.isErr()) {
            log.error(`An error occurred execute Starlark package '${packageRootPath}'`);
            throw runResult.error
        }

        expect(runResult.value.interpretationError).not.toBeUndefined()
        expect(runResult.value.interpretationError?.getErrorMessage()).toContain("Evaluation error: key \"greetings\" not in dict")
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()

        expect(runResult.value.runOutput).toEqual("")
        expect(runResult.value.instructions).toHaveLength(0)
    } finally {
        stopEnclaveFunction()
    }
})
