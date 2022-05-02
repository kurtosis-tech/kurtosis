import {err, ok, Result, Err} from "neverthrow";
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    ContainerConfig,
    ContainerConfigBuilder,
    FilesArtifactID,
    PortProtocol,
    PortSpec,
    ServiceID
} from "kurtosis-core-api-lib";
import axios from "axios";
import log from "loglevel";

const ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN = "upload-test-typescript-"
const ARCHIVE_SUBDIRECTORY_TEST_PATTERN = "sub-folder-"
const ARCHIVE_TEST_FILE_PATTERN = "test-file-"
const ARCHIVE_TEST_FILE_EXTENSION = ".txt"
const ARCHIVE_TEST_CONTENT = "This file is for testing purposes."

const NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY = 3
const NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY = 1

const ENCLAVE_TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = true
const TEST_SERVICE = "UPLOAD_FILES_TEST_SERVICE"
const USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT = "/static"

const FILE_SERVER_PORT_ID = "http"
const FILE_SERVER_PRIVATE_PORT_NUM = 80
const FILE_SERVER_PORT_SPEC = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )
const FILE_SERVER_SERVICE_IMAGE = "flashspys/nginx-static"
const FILE_SERVER_SERVICE_ID: ServiceID = "file-server"

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES = 15
const WAIT_INITIAL_DELAY_MILLISECONDS = 0

jest.setTimeout(180000)

test("Test Upload Files", TestUploadFiles)

async function TestUploadFiles() {
    const tempDirectory = await createTestFolderToUpload()
    if (tempDirectory.isErr()) { throw tempDirectory.error }
    const pathToUpload = tempDirectory.value

    const createEnclaveResult = await createEnclave(ENCLAVE_TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const enclaveCtx = createEnclaveResult.value.enclaveContext

    try {
        const uploadResults = await enclaveCtx.uploadFiles(pathToUpload)
        if(uploadResults.isErr()) { throw uploadResults.error }
        const filesArtifactId = uploadResults.value

        const filesArtifactsMountpoints = new Map<FilesArtifactID, string>()
        filesArtifactsMountpoints.set(filesArtifactId, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT)

        const fileServerContainerConfigSupplier = getFileServerContainerConfigSupplier(filesArtifactsMountpoints)

        const addServiceResult = await enclaveCtx.addService(FILE_SERVER_SERVICE_ID, fileServerContainerConfigSupplier)
        if(addServiceResult.isErr()){ throw addServiceResult.error }

        const serviceContext = addServiceResult.value
        const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
        if(publicPort === undefined){
            throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
        }

        const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
        const fileServerPublicPortNum = publicPort.number
        const filename = path.join(filesArtifactId)

        const waitForHttpGetEndpointAvailabilityResult = await enclaveCtx.waitForHttpGetEndpointAvailability(
            FILE_SERVER_SERVICE_ID,
            FILE_SERVER_PRIVATE_PORT_NUM,
            filename,
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

        const fileRetrievalResults = await getFileContents(
            fileServerPublicIp,
            fileServerPublicPortNum,
            filename
        )
        if(fileRetrievalResults.isErr()){
            log.error("An error occurred getting file 1's contents")
            throw fileRetrievalResults.error
        }

        fileRetrievalResults.value
    } catch (err) {
        throw err
    }
    jest.clearAllTimers()
}

//========================================================================
// Helpers
//========================================================================
async function createTestFiles(directory : string, fileCount : number): Promise<Result<null, Error>>{
    for (let i = 0; i < fileCount; i++) {
        const filename = `${ARCHIVE_TEST_FILE_PATTERN}${i}${ARCHIVE_TEST_FILE_EXTENSION}`
        const fullFilepath = path.join(directory, filename)
        try {
            await filesystem.promises.writeFile(fullFilepath, ARCHIVE_TEST_CONTENT)
        } catch {
            return err(new Error(`Failed to create a test file at '${fullFilepath}'.`))
        }
    }
    return ok(null)
}

async function createTestFolderToUpload(): Promise<Result<string, Error>> {
    const testDirectory = os.tmpdir()

    //Create a base directory
    const tempDirectoryResult = await createTempDirectory(testDirectory, ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN)
    if(tempDirectoryResult.isErr()) { return err(tempDirectoryResult.error) }

    //Create a single subdirectory.
    const subDirectoryResult = await  createTempDirectory(tempDirectoryResult.value, ARCHIVE_SUBDIRECTORY_TEST_PATTERN)
    if(subDirectoryResult.isErr()) { return err(subDirectoryResult.error) }

    //Create NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY
    const subdirTestFileResult = await createTestFiles(subDirectoryResult.value, NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY)
    if(subdirTestFileResult.isErr()){ return err(subdirTestFileResult.error)}

    //Create NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY
    const rootdirTestFileResults = await createTestFiles(testDirectory, NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY)
    if(rootdirTestFileResults.isErr()) { return err(rootdirTestFileResults.error) }

    return ok(testDirectory)
}

async function createTempDirectory(directoryBase: string, directoryPattern: string): Promise<Result<string, Error>> {
    const tempDirpathPrefix = path.join(directoryBase, directoryPattern)
    const tempDirpathResult = await filesystem.promises.mkdtemp(
        tempDirpathPrefix,
    ).then((folder: string) => {
        return ok(folder)
    }).catch((tempDirErr: Error) => {
        return err(tempDirErr)
    });

    if (tempDirpathResult.isErr()) {
        return err(new Error("Failed to create temporary directory for 'uploadFiles' testing."))
    }
    return ok(tempDirpathResult.value)
}

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

async function getFileContents(ipAddress: string, portNum: number, filename: string): Promise<Result<any, Error>> {
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