import { ContainerConfig, ContainerConfigBuilder, ServiceContext } from "kurtosis-core-api-lib";
import log from "loglevel";
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "exec-command";
const IS_PARTITIONING_ENABLED = false;
const EXEC_CMD_TEST_IMAGE = "alpine:3.12.4";
const INPUT_FOR_LOG_OUTPUT_TEST = "hello";
const EXPECTED_LOG_OUTPUT = "hello\n";
const INPUT_FOR_ADVANCED_LOG_OUTPUT_TEST = "hello && hello";
const EXPECTED_ADVANCED_LOG_OUTPUT = "hello && hello\n"
const TEST_SERVICE_ID = "test";
const SUCCESS_EXIT_CODE = 0;

const EXEC_COMMAND_THAT_SHOULD_WORK = ["true"]
const EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT = ["echo", INPUT_FOR_LOG_OUTPUT_TEST]
// This command tests to ensure that the commands the user is running get passed exactly as-is to the Docker
// container. If Kurtosis code is magically wrapping the code with "sh -c", this will fail.
const EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED = [
    "echo",
    INPUT_FOR_ADVANCED_LOG_OUTPUT_TEST,
]
const EXEC_COMMAND_THAT_SHOULD_FAIL = ["false"]

jest.setTimeout(180000)

test("Test exec command", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    
    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfig = getContainerConfig()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfig)

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
        log.info("Exec command returning an error exited with error");

        log.info(`Running exec command '${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}' that should return log output...`)
        const runExecCmdShouldLogOutputResult = await runExecCmd(testServiceContext, EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT)
        if(runExecCmdShouldLogOutputResult.isErr()){
            log.error(`An error occurred running exec command '${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}'`)
            throw runExecCmdShouldLogOutputResult.error
        }
        const [ shouldHaveLogOutputExitCode, logOutput ] = runExecCmdShouldLogOutputResult.value;
        if(SUCCESS_EXIT_CODE !== shouldHaveLogOutputExitCode){
            throw new Error(`Exec command '${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}' should work, but got unsuccessful exit code ${shouldHaveLogOutputExitCode}`)
        }
        if(EXPECTED_LOG_OUTPUT !== logOutput){
            throw new Error(`Exec command '${EXEC_COMMAND_THAT_SHOULD_HAVE_LOG_OUTPUT}' should return '${EXPECTED_LOG_OUTPUT}', but got ${logOutput}`)
        }
        log.info("Exec command returning log output passed as expected");

        log.info(`Running exec command '${EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED}' that will fail if Kurtosis is accidentally sh-wrapping the command...`)
        const shouldFailIfShWrappedExecCmdResult = await runExecCmd(testServiceContext, EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED)
        if(shouldFailIfShWrappedExecCmdResult.isErr()){
            log.error(`An error occurred running exec command '${EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED}'`)
            throw shouldFailIfShWrappedExecCmdResult.error
        }
        const [ shouldNotGetShWrappedExitCode, shouldNotGetShWrappedLogOutput ] = shouldFailIfShWrappedExecCmdResult.value;
        if(SUCCESS_EXIT_CODE !== shouldNotGetShWrappedExitCode){
            throw new Error(`Exec command '${EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED}' should work, but got unsuccessful exit code ${shouldNotGetShWrappedExitCode}`)
        }
        if(EXPECTED_ADVANCED_LOG_OUTPUT !== shouldNotGetShWrappedLogOutput){
            throw new Error(`Exec command '${EXEC_COMMAND_THAT_WILL_FAIL_IF_SH_WRAPPED}' should return '${EXPECTED_ADVANCED_LOG_OUTPUT}', but got ${shouldNotGetShWrappedLogOutput}`)
        }
        log.info("Exec command that will fail if Kurtosis is accidentally sh-wrapping did not fail");

    }finally{
        stopEnclaveFunction()
    }
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfig(): ContainerConfig {
    const entrypointArgs = ["sleep"]
    const cmdArgs = ["30"]

    const containerConfig = new ContainerConfigBuilder(EXEC_CMD_TEST_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .build()

    return containerConfig
}

async function runExecCmd(serviceContext: ServiceContext, command: string[]) {
    const execCommandResult = await serviceContext.execCommand(command)
    if(execCommandResult.isErr()) {
        return err(execCommandResult.error)
    }
    return ok(execCommandResult.value)
}