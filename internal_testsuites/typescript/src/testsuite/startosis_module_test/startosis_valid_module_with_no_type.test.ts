import log from "loglevel";

import {createEnclave} from "../../test_helpers/enclave_setup";
import * as path from "path";
import {DEFAULT_DRY_RUN, EMPTY_EXECUTE_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import {generateScriptOutput} from "../../test_helpers/startosis_helpers";

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

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if (executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Hello World!\n"

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual(expectedScriptOutput)

        expect(executeStartosisModuleValue.getInterpretationError()).toBeUndefined()
        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()
    } finally {
        stopEnclaveFunction()
    }
})

