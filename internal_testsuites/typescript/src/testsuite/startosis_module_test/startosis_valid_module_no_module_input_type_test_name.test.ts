import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    EMPTY_EXECUTE_PARAMS,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with no module input type in types file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if (executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Hello world!\n"

        if (expectedScriptOutput !== executeStartosisModuleValue.getSerializedScriptOutput()) {
            throw err(new Error(`Expected output to be '${expectedScriptOutput} got '${executeStartosisModuleValue.getSerializedScriptOutput()}'`))
        }

        if (executeStartosisModuleValue.getInterpretationError() !== "") {
            throw err(new Error(`Expected Empty Interpretation Error got '${executeStartosisModuleValue.getInterpretationError()}'`))
        }

        if (executeStartosisModuleValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisModuleValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisModuleValue.getValidationErrorsList()}'`))
        }

    } finally {
        stopEnclaveFunction()
    }
})
