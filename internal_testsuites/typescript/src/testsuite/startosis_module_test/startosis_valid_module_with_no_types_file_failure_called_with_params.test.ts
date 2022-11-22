import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

const VALID_MODULE_NO_TYPE_TEST_NAME = "valid-module-no-type";
const VALID_MODULE_NO_TYPE_REL_PATH = "../../../../startosis/valid-kurtosis-module-no-type"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test valid startosis module with no type - failure called with params", async () => {
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

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if (executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;
        if (executeStartosisModuleValue.getInterpretationError() === undefined) {
            throw err(new Error("Expected interpretation errors but got empty interpretation errors"))
        }

        if (!executeStartosisModuleValue.getInterpretationError()?.getErrorMessage().includes("A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid).")) {
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
    } finally {
        stopEnclaveFunction()
    }
})
