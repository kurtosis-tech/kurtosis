import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_RUN_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

const MISSING_MAIN_FUNCTION_TEST_NAME = "invalid-package-missing-main"
const PACKAGE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH = "../../../../starlark/no-run-in-main-star"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid package with no main in main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_FUNCTION_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, PACKAGE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const runResult = await enclaveContext.runStarlarkPackageBlocking(packageRootPath, EMPTY_RUN_PARAMS, DEFAULT_DRY_RUN)
        if (runResult.isErr()) {
            throw err(new Error("Unexpected execution error"))
        }

        expect(runResult.value.interpretationError).not.toBeUndefined()
        expect(runResult.value.interpretationError?.getErrorMessage())
            .toContain("No 'run' function found in file 'github.com/sample/sample-kurtosis-package/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package")

        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()
        expect(runResult.value.runOutput).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
