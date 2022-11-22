import log from "loglevel";
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import * as path from "path";
import {generateScriptOutput} from "../../test_helpers/startosis_helpers";

const IS_PARTITIONING_ENABLED = false;
const EMPTY_EXECUTE_PARAMS = "{}"
const DEFAULT_DRY_RUN = false;

const VALID_MODULE_WITH_TYPES_TEST_NAME = "valid-module-with-types";
const VALID_MODULE_WITH_TYPES_REL_PATH = "../../../../startosis/valid-kurtosis-module-with-types"

const VALID_MODULE_NO_TYPE_TEST_NAME = "valid-module-no-type";
const VALID_MODULE_NO_TYPE_REL_PATH = "../../../../startosis/valid-kurtosis-module-no-type"

const VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME = "valid-module-no-input-type";
const VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH = "../../../../startosis/valid-kurtosis-module-no-module-input-type";

const INVALID_TYPES_FILE_TEST_NAME = "invalid-types-file"
const INVALID_TYPES_FILE_REL_PATH = "../../../../startosis/invalid-types-file"

const INVALID_MODULE_NO_TYPE_BUT_INPUT_ARGS_TEST_NAME = "invalid-module-no-type-input-args";
const INVALID_MODULE_NO_TYPE_BUT_INPUT_ARGS_REL_PATH = "../../../../startosis/invalid-no-type-but-input-args";

const INVALID_KURTOSIS_MOD_TEST_NAME = "invalid-module-invalid-mod-file"
const INVALID_KURTOSIS_MOD_IN_MODULE_REL_PATH = "../../../../startosis/invalid-mod-file"

const MISSING_MAIN_STAR_TEST_NAME = "invalid-module-no-main-file"
const MODULE_WITH_NO_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-star"

const MISSING_MAIN_FUNCTION_TEST_NAME = "invalid-module-missing-main"
const MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-in-main-star"


jest.setTimeout(180000)

test("Test valid startosis module with types", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_WITH_TYPES_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_WITH_TYPES_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Bonjour!\nHello World!\n"

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual(expectedScriptOutput)

        expect(executeStartosisModuleValue.getInterpretationError()).toBeUndefined()
        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()
    }finally{
        stopEnclaveFunction()
    }
})

test("Test valid startosis module with no type", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Hello World!\n"

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual(expectedScriptOutput)

        expect(executeStartosisModuleValue.getInterpretationError()).toBeUndefined()
        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()
    }finally{
        stopEnclaveFunction()
    }
})

test("Test valid startosis module with no type - failure called with params", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
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
    }finally{
        stopEnclaveFunction()
    }
})

test("Test valid startosis module with no module input type in types file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(VALID_MODULE_NO_MODULE_INPUT_TYPE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_MODULE_NO_MODULE_INPUT_TYPE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        const expectedScriptOutput = "Hello world!\n"

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual(expectedScriptOutput)

        expect(executeStartosisModuleValue.getInterpretationError()).toBeUndefined()
        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()
    }finally{
        stopEnclaveFunction()
    }
})

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

        expect(executeStartosisModuleValue.getInterpretationError()).not.toBeUndefined()
        expect(executeStartosisModuleValue.getInterpretationError()?.getErrorMessage())
            .toContain("A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but 'ModuleInput' type is not defined in the module's 'types.proto' file.")

        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual("")
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid startosis module invalid types file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(INVALID_TYPES_FILE_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, INVALID_TYPES_FILE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const serializedParams = "{\"greetings\": \"Bonjour!\"}"
        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, serializedParams, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
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
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid startosis module no types file but input_args in main", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(INVALID_MODULE_NO_TYPE_BUT_INPUT_ARGS_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, INVALID_MODULE_NO_TYPE_BUT_INPUT_ARGS_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            log.error(`An error occurred execute startosis module '${moduleRootPath}'`);
            throw executeStartosisModuleResult.error
        }
        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        expect(executeStartosisModuleValue.getInterpretationError()).not.toBeUndefined()
        expect(executeStartosisModuleValue.getInterpretationError()?.getErrorMessage())
            .toContain("Evaluation error: function main missing 1 argument (input_args)")

        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual("")
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid module with invalid mod file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(INVALID_KURTOSIS_MOD_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, INVALID_KURTOSIS_MOD_IN_MODULE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(!executeStartosisModuleResult.isErr()) {
            throw err(new Error("Module with invalid module was expected to error but didn't"))
        }

        if (!executeStartosisModuleResult.error.message.includes(`Field module.name in kurtosis.mod needs to be set and cannot be empty`)) {
            throw err(new Error("Unexpected error message"))
        }
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid module with no main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_STAR_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(!executeStartosisModuleResult.isErr()) {
            throw err(new Error("Module with invalid module was expected to error but didn't"))
        }

        expect(executeStartosisModuleResult.error.message).toContain("An error occurred while verifying that 'main.star' exists on root of module")
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid module with no main in main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(MISSING_MAIN_FUNCTION_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath, EMPTY_EXECUTE_PARAMS, DEFAULT_DRY_RUN)

        if(executeStartosisModuleResult.isErr()) {
            throw err(new Error("Unexpected execution error"))
        }

        const executeStartosisModuleValue = executeStartosisModuleResult.value;

        expect(executeStartosisModuleValue.getInterpretationError()).not.toBeUndefined()
        expect(executeStartosisModuleValue.getInterpretationError()?.getErrorMessage())
            .toContain("Evaluation error: module has no .main field or method\n\tat [3:12]: <toplevel>")

        expect(executeStartosisModuleValue.getExecutionError()).toBeUndefined()
        expect(executeStartosisModuleValue.getValidationErrors()).toBeUndefined()

        expect(generateScriptOutput(executeStartosisModuleValue.getKurtosisInstructionsList())).toEqual("")
    }finally{
        stopEnclaveFunction()
    }
})
