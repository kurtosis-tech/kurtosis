import { ContainerConfig, ContainerConfigBuilder, FilesArtifactID, PortProtocol, PortSpec, ServiceID } from "kurtosis-core-api-lib"
import log from "loglevel";
import { Result, ok, err } from "neverthrow";
import axios from "axios"

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "files-artifact-mounting"
const IS_PARTITIONING_ENABLED = false

const FILE_SERVER_SERVICE_IMAGE = "flashspys/nginx-static"
const FILE_SERVER_SERVICE_ID: ServiceID = "file-server"
const SECOND_FILE_SERVER_SERVICE_ID: ServiceID = "second-file-server"
const FILE_SERVER_PORT_ID = "http"
const FILE_SERVER_PRIVATE_PORT_NUM = 80

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES = 15
const WAIT_INITIAL_DELAY_MILLISECONDS = 0

const TEST_FILES_ARTIFACT_ID: FilesArtifactID = "test-files-artifact"
const SECOND_TEST_FILES_ARTIFACT_ID: FilesArtifactID = "second-test-files-artifact"
const TEST_FILES_ARTIFACT_URL = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

// Filenames & contents for the files stored in the files artifact
const FILE1_FILENAME = "file1.txt"
const FILE2_FILENAME = "file2.txt"

const EXPECTED_FILE1_CONTENTS = "file1\n"
const EXPECTED_FILE2_CONTENTS = "file2\n"

const FILE_SERVER_PORT_SPEC = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )

const USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT  = "/static"

const DUPLICATE_MOUNTPOINT_DOCKER_DAEMON_ERR_MSG  = "Duplicate mount point"

jest.setTimeout(180000)

test("Test web file storing", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {

        // ------------------------------------- TEST SETUP ----------------------------------------------
        const filesArtifacts = new Map<FilesArtifactID, string>()
        filesArtifacts.set(TEST_FILES_ARTIFACT_ID, TEST_FILES_ARTIFACT_URL)
        filesArtifacts.set(SECOND_TEST_FILES_ARTIFACT_ID, TEST_FILES_ARTIFACT_URL)
        const storeWebFilesResult = await enclaveContext.storeWebFiles(TEST_FILES_ARTIFACT_URL);
        if(storeWebFilesResult.isErr()) { throw storeWebFilesResult.error }
        const filesArtifactId = storeWebFilesResult.value;

        const filesArtifactsMountpoints = new Map<FilesArtifactID, string>()
        filesArtifactsMountpoints.set(filesArtifactId, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT)

        const fileServerContainerConfigSupplier = getFileServerContainerConfigSupplier(filesArtifactsMountpoints)

        const addServiceResult = await enclaveContext.addService(FILE_SERVER_SERVICE_ID, fileServerContainerConfigSupplier)

        if(addServiceResult.isErr()){ throw addServiceResult.error }

        const serviceContext = addServiceResult.value
        const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
        if(publicPort === undefined){
            throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
        }

        const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
        const fileServerPublicPortNum = publicPort.number

        // TODO It's suuuuuuuuuuper confusing that we have to pass the private port in here!!!! We should just require the user
        //  to pass in the port ID and the API container will translate that to the private port automatically!!!
        const waitForHttpGetEndpointAvailabilityResult = await enclaveContext.waitForHttpGetEndpointAvailability(
            FILE_SERVER_SERVICE_ID, 
            FILE_SERVER_PRIVATE_PORT_NUM,
            FILE1_FILENAME, 
            WAIT_INITIAL_DELAY_MILLISECONDS, 
            WAIT_FOR_STARTUP_MAX_RETRIES, 
            WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, 
            ""
        );

        if(waitForHttpGetEndpointAvailabilityResult.isErr()){
            log.error("An error occurred waiting for the file server service to become available")
            throw waitForHttpGetEndpointAvailabilityResult.error
        }

        log.info(`Added file server service with public IP "${fileServerPublicIp}" and port "${fileServerPublicPortNum}"`)

        // ------------------------------------- TEST RUN ----------------------------------------------

        const file1ContentsResult = await getFileContents(
            fileServerPublicIp,
            fileServerPublicPortNum,
            FILE1_FILENAME
        )
        if(file1ContentsResult.isErr()){
            log.error("An error occurred getting file 1's contents")
            throw file1ContentsResult.error
        }

        const file1Contents = file1ContentsResult.value
        if(file1Contents !== EXPECTED_FILE1_CONTENTS){
            throw new Error(`Actual file 1 contents "${file1Contents}" != expected file 1 contents "${EXPECTED_FILE1_CONTENTS}"`)
        }

        const file2ContentsResult = await getFileContents(
            fileServerPublicIp,
            fileServerPublicPortNum,
            FILE2_FILENAME
        )

        if(file2ContentsResult.isErr()){
            log.error("An error occurred getting file 2's contents")
            throw file2ContentsResult.error
        }

        const file2Contents = file2ContentsResult.value
        if(file2Contents !== EXPECTED_FILE2_CONTENTS){
            throw new Error(`Actual file 2 contents "${file2Contents}" != expected file 2 contents "${EXPECTED_FILE2_CONTENTS}"`)
        }

        const duplicateFilesArtifactMountpoints = new Map<FilesArtifactID, string>()
        duplicateFilesArtifactMountpoints.set(TEST_FILES_ARTIFACT_ID, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT)
        duplicateFilesArtifactMountpoints.set(SECOND_TEST_FILES_ARTIFACT_ID, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT)

        //TODO the error is detected in Docker, it is enough for now, but we should capture it in Kurt Core for optimization and decoupling
        const wrongFileServerContainerConfigSupplier = getFileServerContainerConfigSupplier(duplicateFilesArtifactMountpoints)

        const addSecondServiceResult = await enclaveContext.addService(SECOND_FILE_SERVER_SERVICE_ID, wrongFileServerContainerConfigSupplier)

        if(addSecondServiceResult.isOk()){
            throw new Error(`Adding service "${SECOND_FILE_SERVER_SERVICE_ID}" did not fails and it is wrong, because the files artifact mountpoints set in the container config supplier are duplicate and it must not be allowed`)
        }

        if(addSecondServiceResult.isErr()){
            const errMsg = addSecondServiceResult.error.message
            if(!errMsg.includes(DUPLICATE_MOUNTPOINT_DOCKER_DAEMON_ERR_MSG)){
               throw new Error(`Adding service "${SECOND_FILE_SERVER_SERVICE_ID}" has failed, but the error is not the duplicated-files-artifact-mountpoints-error that we expected, this is throwing this error instead:\n "${errMsg}"`)
            }
        }

    }finally{
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

function getFileServerContainerConfigSupplier(filesArtifactMountpoints: Map<FilesArtifactID, string>): (ipAddr: string) => Result<ContainerConfig, Error> {
	
    const containerConfigSupplier = (ipAddr:string): Result<ContainerConfig, Error> => {

        const usedPorts = new Map<string, PortSpec>()
        usedPorts.set(FILE_SERVER_PORT_ID, FILE_SERVER_PORT_SPEC)

        const containerConfig = new ContainerConfigBuilder(FILE_SERVER_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withFiles(filesArtifactMountpoints)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier
}

async function getFileContents(ipAddress: string, portNum: number, filename: string): Promise<Result<string, Error>> {
    let response;
    try {
        response = await axios(`http://${ipAddress}:${portNum}/${filename}`)
    }catch(error){
        log.error(`An error occurred getting the contents of file "${filename}"`)
        if(error instanceof Error){
            return err(error)
        }else{
            return err(new Error("An error occurred getting the contents of file, but the error wasn't of type Error"))
        }
    }
    const bodyStr = String(response.data)
    return ok(bodyStr)
}