import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import log from "loglevel";
import * as path from "path";
import {generateScriptOutput} from "../../test_helpers/startosis_helpers";

const INVALID_TYPES_FILE_TEST_NAME = "invalid-types-file"
const INVALID_TYPES_FILE_REL_PATH = "../../../../startosis/invalid-types-file"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid startosis module invalid types file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(INVALID_TYPES_FILE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, INVALID_TYPES_FILE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if (executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        expect(executeStartosisModuleValue.getInterpretationError()).not.toBeUndefined()
        expect(executeStartosisModuleValue.getInterpretationError()?.getErrorMessage())
            .toContain("A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid).")

        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual("")
    } finally {
        stopEnclaveFunction()
    }
})
