import { ContainerConfig, ContainerConfigBuilder, ServiceContext, SharedPath } from "kurtosis-core-api-lib";
import log from "loglevel";
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "exec-command";
const IS_PARTITIONING_ENABLED = false;
const EXEC_CMD_TEST_IMAGE = "alpine:3.12.4";
const INPUT_FOR_LOG_OUTPUT_TEST = "hello";
const EXPECTED_LOG_OUTPUT = "hello\n";
const TEST_SERVICE_ID = "test";
const SUCCESS_EXIT_CODE = 0;

const EXEC_COMMAND_THAT_SHOULD_WORK = ["true"]
const EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT = ["echo", INPUT_FOR_LOG_OUTPUT_TEST]
const EXEC_COMMAND_THAT_SHOULD_FAIL = ["false"]

jest.setTimeout(180000)

test("Test exec command", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    
    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfigSupplier = getContainerConfigSupplier()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfigSupplier)

        if(addServiceResult.isErr()) {
            log.error(`An error occurred starting service "${TEST_SERVICE_ID}"`);
            throw addServiceResult.error
        };

        const testServiceContext = addServiceResult.value

        // ------------------------------------- TEST RUN ----------------------------------------------
        log.info(`Running exec command "${EXEC_COMMAND_THAT_SHOULD_WORK}" that should return a successful exit code...`)
        
        const runExecCmdShouldWorkResult = await runExecCmd(testServiceContext, EXEC_COMMAND_THAT_SHOULD_WORK)

        if(runExecCmdShouldWorkResult.isErr()){
            log.error(`An error occurred running exec command "${EXEC_COMMAND_THAT_SHOULD_WORK}"`)
            throw runExecCmdShouldWorkResult.error
        }
        const [ shouldWorkExitCode ] = runExecCmdShouldWorkResult.value;

        if(SUCCESS_EXIT_CODE !== shouldWorkExitCode){
            throw new Error(`Exec command "${EXEC_COMMAND_THAT_SHOULD_WORK}" should work, but got unsuccessful exit code ${shouldWorkExitCode}`)
        }

        log.info("Exec command returned successful exit code as expected")

        log.info(`Running exec command "${EXEC_COMMAND_THAT_SHOULD_FAIL}" that should return an error exit code...`)

        const runExecCmdShouldFailResult = await runExecCmd(testServiceContext, EXEC_COMMAND_THAT_SHOULD_FAIL)

        if(runExecCmdShouldFailResult.isErr()){
            log.error(`An error occurred running exec command "${EXEC_COMMAND_THAT_SHOULD_FAIL}"`)
            throw runExecCmdShouldFailResult.error
        }
        const [ shouldFailExitCode ] = runExecCmdShouldFailResult.value;

        if(SUCCESS_EXIT_CODE === shouldFailExitCode){
            throw new Error(`Exec command "${EXEC_COMMAND_THAT_SHOULD_FAIL}" should fail, but got successful exit code ${SUCCESS_EXIT_CODE}`)
        }

        log.info(`Running exec command "${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}" that should return log output...`)

        const runExecCmdShouldLogOutputResult = await runExecCmd(testServiceContext, EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT)

        if(runExecCmdShouldLogOutputResult.isErr()){
            log.error(`An error occurred running exec command "${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}"`)
            throw runExecCmdShouldLogOutputResult.error
        }
        const [ shouldHaveLogOutputExitCode, logOutput ] = runExecCmdShouldLogOutputResult.value;

        if(SUCCESS_EXIT_CODE !== shouldHaveLogOutputExitCode){
            throw new Error(`Exec command "${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}" should work, but got unsuccessful exit code ${shouldHaveLogOutputExitCode}`)
        }

        if(EXPECTED_LOG_OUTPUT !== logOutput){
            throw new Error(`Exec command "${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}" should return ${INPUT_FOR_LOG_OUTPUT_TEST}, but got ${logOutput}`)
        }
        
        log.info("Exec command returned error exit code as expected")

    }finally{
        stopEnclaveFunction()
    }
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfigSupplier(): (ipAddr:string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {
	
    const containerConfigSupplier = (ipAddr:string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {
        const entrypointArgs = ["sleep"]
        const cmdArgs = ["30"]

        const containerConfig = new ContainerConfigBuilder(EXEC_CMD_TEST_IMAGE)
            .withEntrypointOverride(entrypointArgs)
            .withCmdOverride(cmdArgs)
            .build()
        
        return ok(containerConfig)
    }

    return containerConfigSupplier
}

async function runExecCmd(serviceContext: ServiceContext, command: string[]) {
    const execCommandResult = await serviceContext.execCommand(command)
    if(execCommandResult.isErr()) {
        return err(execCommandResult.error)
    }
    return ok(execCommandResult.value)
}