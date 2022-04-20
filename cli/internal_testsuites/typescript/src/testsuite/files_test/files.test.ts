import { ContainerConfig, ContainerConfigBuilder, ServiceID, SharedPath } from "kurtosis-core-api-lib";
import log from "loglevel"
import { ok, Result, err } from "neverthrow";
import * as fs from "fs"
import * as utf8 from "utf8"

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "files";
const IS_PARTITIONING_ENABLED = false;

const DOCKER_IMAGE = "alpine:3.12.4"
const TEST_SERVICE:ServiceID = "test-service"

const EXEC_COMMAND_SUCCESS_EXIT_CODE = 0
const EXPECTED_TEST_FILE1_CONTENTS = "This is a test file"
const EXPECTED_TEST_FILE2_CONTENTS = "This is another test file"
const GENERATED_FILE_PERM_BITS = 0o644

// Mapping of filepath_rel_to_shared_dir_root -> contents
const generatedFileRelPathsAndContents = new Map<string,string>();
generatedFileRelPathsAndContents.set("test1.txt", EXPECTED_TEST_FILE1_CONTENTS)
generatedFileRelPathsAndContents.set("test2.txt", EXPECTED_TEST_FILE2_CONTENTS)

jest.setTimeout(180000)

test("Test files", async () => {

    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction, kurtosisContext } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfigSupplier = getContainerConfigSupplier()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE, containerConfigSupplier)
        if(addServiceResult.isErr()) {
            log.error("An error occurred adding the file server service") 
            throw addServiceResult.error 
        }

        const serviceContext = addServiceResult.value

	    // ------------------------------------- TEST RUN ----------------------------------------------
        for(let [relativeFilepath, expectedContents] of generatedFileRelPathsAndContents) {
            const sharedFilepath = serviceContext.getSharedDirectory().getChildPath(relativeFilepath)

            const catStaticFileCmd = ["cat", sharedFilepath.getAbsPathOnServiceContainer()]

            const execCommandResult = await serviceContext.execCommand(catStaticFileCmd)

            if(execCommandResult.isErr()){
                log.error(`An error occurred executing command "${catStaticFileCmd}" to cat the static file "${relativeFilepath}" contents`)
                throw execCommandResult.error
            }

            const [ exitCode, actualContents ] = execCommandResult.value

            if(EXEC_COMMAND_SUCCESS_EXIT_CODE !== exitCode){
                throw new Error(`Command "${catStaticFileCmd}" to cat the static file "${relativeFilepath}" exited with non-successful exit code "${exitCode}"`)
            }

            if(expectedContents !== actualContents){
                throw new Error(`Static file contents "${actualContents}" don't match expected static file "${relativeFilepath}" contents "${expectedContents}"`)
            }

            log.info(`Static file "${relativeFilepath}" contents were "${expectedContents}" as expected`)
        }
    
    }finally{
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfigSupplier():(ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = (ipAddr: string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {
        for (const [relativeFilePath, contents] of generatedFileRelPathsAndContents) {
            const generateFileInServiceContainerResult = generateFileInServiceContainer(relativeFilePath, contents, sharedDirectory)
            if(generateFileInServiceContainerResult.isErr()){
                log.error(`An error occurred generating file with relative filepath "${relativeFilePath}"`)
                return err(generateFileInServiceContainerResult.error)
            }
        }

        // We sleep because the only function of this container is to test Docker executing a command while it's running
        // NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
        // args), but this provides a nice little regression test of the ENTRYPOINT overriding
        const entrypointArgs = ["sleep"]
        const cmdArgs = ["30"]

        const containerConfig = new ContainerConfigBuilder(DOCKER_IMAGE)
            .withEntrypointOverride(entrypointArgs)
            .withCmdOverride(cmdArgs)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier
}

function generateFileInServiceContainer(relativePath: string, contents: string, sharedDirectory: SharedPath): Result<null,Error>{
    const sharedFilepath = sharedDirectory.getChildPath(relativePath)
    const absFilepathOnThisContainer = sharedFilepath.getAbsPathOnThisContainer()

    const bytesArray = utf8.encode(contents)

    try {
        fs.writeFileSync(sharedFilepath.getAbsPathOnThisContainer(),bytesArray, { mode: GENERATED_FILE_PERM_BITS })
    }catch(error){
        log.error(`An error occurred writing contents "${contents}" to file "${absFilepathOnThisContainer}" with perms "${GENERATED_FILE_PERM_BITS}"`)
        if(error instanceof Error){
            return err(error)
        }else{
            return err(new Error("Encountered error while writing the file, but the error wasn't of type Error"))
        }
    }

    return ok(null)
};