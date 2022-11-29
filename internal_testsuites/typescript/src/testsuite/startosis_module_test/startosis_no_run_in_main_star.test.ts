import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_RUN_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";
import {readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

const MISSING_MAIN_FUNCTION_TEST_NAME = "invalid-module-missing-main"
const MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH = "../../../../startosis/no-run-in-main-star"

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
        const packageRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const outputStream = await enclaveContext.runStarlarkPackage(packageRootPath, EMPTY_RUN_PARAMS, DEFAULT_DRY_RUN)

        if (outputStream.isErr()) {
            throw err(new Error("Unexpected execution error"))
        }

        const [scriptOutput, _, interpretationError, validationErrors, executionError] = await readStreamContentUntilClosed(outputStream.value);

        expect(interpretationError).not.toBeUndefined()
        expect(interpretationError?.getErrorMessage())
            .toContain("No 'run' function found in file 'github.com/sample/sample-kurtosis-module/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package")

        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()

        expect(scriptOutput).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
