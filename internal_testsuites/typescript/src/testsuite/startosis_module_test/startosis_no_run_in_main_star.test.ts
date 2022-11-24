import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_EXECUTE_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";
import {generateScriptOutput} from "../../test_helpers/startosis_helpers";

const MISSING_MAIN_FUNCTION_TEST_NAME = "invalid-module-missing-main"
const MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-in-main-star"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid module with no main in main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_FUNCTION_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if (executeStartosisModuleResult.isErr()) {
            throw err(new Error("Unexpected execution error"))
        }

        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        expect(executeStartosisModuleValue.getInterpretationError()).not.toBeUndefined()
        expect(executeStartosisModuleValue.getInterpretationError()?.getErrorMessage())
            .toContain("Evaluation error: module has no .run field or method\n\tat [3:12]: <toplevel>")

        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
