import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_EXECUTE_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

const MISSING_MAIN_STAR_TEST_NAME = "invalid-module-no-main-file"
const MODULE_WITH_NO_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-star"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid module with no main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_STAR_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if (!executeStartosisModuleResult.isErr()) {
            throw err(new Error("Module with invalid module was expected to error but didn't"))
        }

        expect(executeStartosisModuleResult.error.message).toContain("An error occurred while verifying that 'main.star' exists on root of module")
    } finally {
        stopEnclaveFunction()
    }
})