import log from "loglevel";
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import * as path from "path";

const TEST_NAME = "startosis-module-command";
const IS_PARTITIONING_ENABLED = false;
const VALID_KURTOSIS_MODULE_REL_PATH = "../../../../startosis/valid-kurtosis-module"
const INVALID_KURTOSIS_MOD_IN_MODULE_REL_PATH = "../../../../startosis/invalid-mod-file"
const MODULE_WITH_NO_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-star"
const MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH = "../../../../startosis/no-main-in-main-star"


jest.setTimeout(180000)

test("Test valid startosis module execution", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, VALID_KURTOSIS_MODULE_REL_PATH)

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

        if (executeStartosisModuleValue.getInterpretationError() !== "") {
            throw err(new Error(`Expected Empty Interpretation Error got '${executeStartosisModuleValue.getInterpretationError()}'`))
        }

        if (executeStartosisModuleValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisModuleValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisModuleValue.getValidationErrorsList()}'`))
        }

    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid module with invalid mod file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, INVALID_KURTOSIS_MOD_IN_MODULE_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath)

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
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath)

        if(!executeStartosisModuleResult.isErr()) {
            throw err(new Error("Module with invalid module was expected to error but didn't"))
        }

        if (!executeStartosisModuleResult.error.message.includes(`An error occurred while verifying that 'main.star' exists on root of module`)) {
            throw err(new Error("Unexpected error message"))
        }
    }finally{
        stopEnclaveFunction()
    }
})

test("Test invalid module with no main in main.star", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const moduleRootPath = path.join(__dirname, MODULE_WITH_NO_MAIN_IN_MAIN_STAR_REL_PATH)

        log.info(`Loading module at path '${moduleRootPath}'`)

        const executeStartosisModuleResult = await enclaveContext.executeStartosisModule(moduleRootPath)

        if(executeStartosisModuleResult.isErr()) {
            throw err(new Error("Unexpected execution error"))
        }

        const executeStartosisModuleValue = executeStartosisModuleResult.value;
        if (executeStartosisModuleValue.getInterpretationError() === "") {
            throw err(new Error("Expected interpretation errors but got empty interpretation errors"))
        }

        if (!executeStartosisModuleValue.getInterpretationError().includes("Evaluation error: load: name main not found in module github.com/sample/sample-kurtosis-module/main.star")) {
            throw err(new Error("Got interpretation error but got invalid contents"))
        }

        if (executeStartosisModuleValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisModuleValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisModuleValue.getValidationErrorsList()}'`))
        }

        if (executeStartosisModuleValue.getSerializedScriptOutput() != "") {
            throw err(new Error(`Expected output to be empty got '${executeStartosisModuleValue.getSerializedScriptOutput()}'`))
        }
    }finally{
        stopEnclaveFunction()
    }
})
