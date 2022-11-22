import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    DEFAULT_DRY_RUN,
    IS_PARTITIONING_ENABLED,
    JEST_TIMEOUT_MS,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH,
    VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME
} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with no module input type in types file - failure called with params", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;
        if (executeStartosisModuleValue.getInterpretationError() === undefined) {
            throw err(new Error("Expected interpretation errors but got empty interpretation errors"))
        }

        if (!executeStartosisModuleValue.getInterpretationError()?.getErrorMessage().includes("A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but 'ModuleInput' type is not defined in the module's 'types.proto' file.")) {
            throw err(new Error("Got interpretation error but got invalid contents"))
        }

        if (executeStartosisModuleValue.getExecutionError() !== undefined) {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisModuleValue.getValidationErrors() !== undefined) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisModuleValue.getValidationErrors()}'`))
        }

        if (executeStartosisModuleValue.getSerializedScriptOutput() != "") {
            throw err(new Error(`Expected output to be empty got '${executeStartosisModuleValue.getSerializedScriptOutput()}'`))
        }
    }finally{
        stopEnclaveFunction()
    }
})
