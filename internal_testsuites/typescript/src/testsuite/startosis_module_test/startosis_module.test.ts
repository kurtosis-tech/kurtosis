import log from "loglevel";
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import * as path from "path";

const TEST_NAME = "startosis-module-command";
const IS_PARTITIONING_ENABLED = false;
const RELATIVE_MODULE_PATH = "../../../../startosis/sample-kurtosis-module"


jest.setTimeout(180000)

test("Test startosis module execution", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, RELATIVE_MODULE_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Hello World!\n"

        if (expectedScriptOutput !== executeStartosisModuleValue.getSerializedScriptOutput()) {
            throw err(new Error(`Expected output to be '${expectedScriptOutput} got '${executeStartosisModuleValue.getSerializedScriptOutput()}'`))
        }

        if (executeStartosisModuleValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Execution Interpretation Error got '${executeStartosisModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisModuleValue.getInterpretationError() !== "") {
            throw err(new Error(`Expected Empty Interpretation Error got '${executeStartosisModuleValue.getInterpretationError()}'`))
        }
        if (executeStartosisModuleValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisModuleValue.getValidationErrorsList()}'`))
        }

    }finally{
        stopEnclaveFunction()
    }
})