import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_RUN_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";
import {readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

const MISSING_MAIN_STAR_TEST_NAME = "invalid-package-no-main-file"
const PACKAGE_WITH_NO_MAIN_STAR_REL_PATH = "../../../../starlark/no-main-star"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid package with no main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_STAR_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, PACKAGE_WITH_NO_MAIN_STAR_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const outputStream = await enclaveContext.runStarlarkPackage(packageRootPath, EMPTY_RUN_PARAMS, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            throw err(new Error(`An error occurred execute Starlark package '${packageRootPath}'`));
        }
        const [scriptOutput, _, interpretationError, validationErrors, executionError] = await readStreamContentUntilClosed(outputStream.value);

        expect(interpretationError).not.toBeUndefined()
        expect(interpretationError?.getErrorMessage())
            .toContain("An error occurred while verifying that 'main.star' exists on root of package")
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()

        expect(scriptOutput).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
