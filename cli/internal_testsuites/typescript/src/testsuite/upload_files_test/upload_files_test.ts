import {err, ok, Result} from "neverthrow";
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

const ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN   = "upload-test-typescript-"
const ARCHIVE_SUBDIRECTORY_TEST_PATTERN     = "sub-folder-"
const ARCHIVE_TEST_FILE_PATTERN             = "test-file-"
const ARCHIVE_TEST_FILE_EXTENSION           = ".txt"
const ARCHIVE_TEST_CONTENT                  = "This file is for testing purposes."

const NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY      = 3
const NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY    = 1

const ENCLAVE_TEST_NAME                                = "upload-files-test"
const IS_PARTITIONING_ENABLED                          = true
const USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT = "/static"

const FILE_SERVER_PORT_ID               = "http"
const FILE_SERVER_PRIVATE_PORT_NUM      = 80
const FILE_SERVER_PORT_SPEC             = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )
const FILE_SERVER_SERVICE_IMAGE         = "flashspys/nginx-static"
const FILE_SERVER_SERVICE_ID: ServiceID = "file-server"

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES = 15
const WAIT_INITIAL_DELAY_MILLISECONDS = 0

//Keywords for mapping paths for file integrity checking.
const DISK_DIR_KEYWORD          = "diskDir"
const ROOT_DIR_KEYWORD          = "rootDir"
const SUB_DIR_KEYWORD           = "subDir"
const SUB_FILE_PATTERN_KEYWORD  = "subFile"
const ROOT_FILE_PATTERN_KEYWORD = "rootFile"

jest.setTimeout(180000)

test("Test Upload Files", TestUploadFiles)

async function TestUploadFiles() {
    const tempDirectory = await createTestFolderToUpload()
    if (tempDirectory.isErr()) { throw tempDirectory.error }
    const allPaths = tempDirectory.value

    const createEnclaveResult = await createEnclave(ENCLAVE_TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const enclaveCtx = createEnclaveResult.value.enclaveContext

    try {
        const pathToUpload = allPaths.get(DISK_DIR_KEYWORD)
        if (typeof pathToUpload === "undefined") {throw new Error("Failed to store uploadable path in path map.")}
        const uploadResults = await enclaveCtx.uploadFiles(pathToUpload)
        if(uploadResults.isErr()) { throw uploadResults.error }
        const filesArtifactId = uploadResults.value

        const filesArtifactsMountPoints = new Map<FilesArtifactID, string>()
        filesArtifactsMountPoints.set(filesArtifactId, USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT)

        const fileServerContainerConfigSupplier = getFileServerContainerConfigSupplier(filesArtifactsMountPoints)
        const addServiceResult = await enclaveCtx.addService(FILE_SERVER_SERVICE_ID, fileServerContainerConfigSupplier)
        if(addServiceResult.isErr()){ throw addServiceResult.error }

        const serviceContext = addServiceResult.value
        const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
        if(publicPort === undefined){
            throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
        }

        const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
        const fileServerPublicPortNum = publicPort.number
        const firstRootFilename = `${allPaths.get(ROOT_DIR_KEYWORD)}/${ROOT_FILE_PATTERN_KEYWORD}0`

        const waitForHttpGetEndpointAvailabilityResult = await enclaveCtx.waitForHttpGetEndpointAvailability(
            FILE_SERVER_SERVICE_ID,
            FILE_SERVER_PRIVATE_PORT_NUM,
            firstRootFilename,
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

        const allContentTestResults = await testAllContent(allPaths, fileServerPublicIp, fileServerPublicPortNum)
        if(allContentTestResults.isErr()) { throw allContentTestResults.error}
    } catch (err) {
        throw err
    }
    jest.clearAllTimers()
}

//========================================================================
// Helpers
//========================================================================
async function testAllContent(allPaths: Map<string,string>, ipAddress: string, portNum: number):
Promise<Result<null, Error>>{

    const rootDirTestResults = await  testDirectoryFiles(allPaths, ROOT_FILE_PATTERN_KEYWORD,
        NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY, ipAddress, portNum)
    if(rootDirTestResults.isErr()){ return err(rootDirTestResults.error) }

    const subDirTestResults = await testDirectoryFiles(allPaths, SUB_FILE_PATTERN_KEYWORD,
        NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY, ipAddress, portNum)
    if (subDirTestResults.isErr()){ return err(subDirTestResults.error) }

    return ok(null)
}

//Cycle through a directory and check the file contents.
async function testDirectoryFiles(allPaths: Map<string, string>, keywordPattern: string, fileCount: number,
ipAddress: string, portNum: number): Promise<Result<null, Error>> {

    for(let i = 0; i < fileCount; i++){
        let fileKeyword = `${keywordPattern}${i}`
        let filename = allPaths.get(fileKeyword)
        if (typeof filename === "undefined"){
            return err(new Error(`The file for keyword ${fileKeyword} was not mapped in the paths map.`))
        }
        let testContentResults = await testFileContents(ipAddress, portNum, filename)
        if (testContentResults.isErr()) { return  err(testContentResults.error) }
    }

    return ok(null)
}

//Test file contents against the ARCHIVE_TEST_CONTENT string.
async function testFileContents(ipAddress: string, portNum: number, filename: string): Promise<Result<null, Error>> {
    let fileContentResults = await getFileContents(ipAddress, portNum, filename)
    if(fileContentResults.isErr()) { return err(fileContentResults.error)}

    let dataAsString = String(fileContentResults.value)
    if (dataAsString !== ARCHIVE_TEST_CONTENT){
        return err(new Error(`The file contents of '${filename}' do not match the test content.\n
                              ${dataAsString} !== ${ARCHIVE_TEST_CONTENT}`))
    }
    return ok(null)
}

async function createTestFiles(directory : string, fileCount : number): Promise<Result<string[], Error>>{
    let filenames: string[] = []

    for (let i = 0; i < fileCount; i++) {
        const filename = `${ARCHIVE_TEST_FILE_PATTERN}${i}${ARCHIVE_TEST_FILE_EXTENSION}`
        const fullFilepath = path.join(directory, filename)
        filenames.push(filename)
        try {
            await filesystem.promises.writeFile(fullFilepath, ARCHIVE_TEST_CONTENT)
            await filesystem.promises.chmod(fullFilepath, 655)
        } catch {
            return err(new Error(`Failed to create a test file at '${fullFilepath}'.`))
        }
    }
    return ok(filenames)
}

async function createTestFolderToUpload(): Promise<Result<Map<string,string>, Error>> {
    const testDirectory = os.tmpdir()

    //Create a base directory
    const tempDirectoryResult = await createTempDirectory(testDirectory, ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN)
    if(tempDirectoryResult.isErr()) { return err(tempDirectoryResult.error) }

    //Create a single subdirectory.
    const subDirectoryResult = await  createTempDirectory(tempDirectoryResult.value, ARCHIVE_SUBDIRECTORY_TEST_PATTERN)
    if(subDirectoryResult.isErr()) { return err(subDirectoryResult.error) }

    //Create NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY
    const subDirTestFileResult = await createTestFiles(subDirectoryResult.value, NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY)
    if(subDirTestFileResult.isErr()){ return err(subDirTestFileResult.error)}

    //Create NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY
    const rootDirTestFileResults = await createTestFiles(testDirectory, NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY)
    if(rootDirTestFileResults.isErr()) { return err(rootDirTestFileResults.error) }

    //Set folder permissions.
    try {
        await filesystem.promises.chmod(tempDirectoryResult.value, 744) //baseDirectory
        await filesystem.promises.chmod(subDirectoryResult.value, 744) //subdirectory
    } catch {
        return err(new Error("Could not set permissions for root directory or subdirectories while creating test files."))
    }

    let subDirFilenames = subDirTestFileResult.value
    let rootFilenames = rootDirTestFileResults.value
    let rootDir = path.basename(tempDirectoryResult.value)
    let subDir = path.basename(subDirectoryResult.value)

    let allPaths = new Map<string, string>()
    allPaths.set(DISK_DIR_KEYWORD, `${testDirectory}/${rootDir}`)
    allPaths.set(ROOT_DIR_KEYWORD, rootDir)
    allPaths.set(SUB_DIR_KEYWORD, subDir)

    for(let i = 0; i < subDirFilenames.length; i++){
        let basename = path.basename(tempDirectoryResult.value)
        let relativeSubFile = `${rootDir}/${subDir}/${basename}`
        let keyword = `${SUB_FILE_PATTERN_KEYWORD}${i}`
        allPaths.set(keyword, relativeSubFile)
    }

    for(let i = 0; i < rootFilenames.length; i++){
        let basename = path.basename(tempDirectoryResult.value)
        let relativeRootFile = `${rootDir}/${basename}`
        let keyword = `${ROOT_FILE_PATTERN_KEYWORD}${i}`
        allPaths.set(keyword, relativeRootFile)
    }

    return ok(allPaths)
}

async function createTempDirectory(directoryBase: string, directoryPattern: string): Promise<Result<string, Error>> {
    const tempDirPathPrefix = path.join(directoryBase, directoryPattern)
    const tempDirPathResult = await filesystem.promises.mkdtemp(
        tempDirPathPrefix,
    ).then((folder: string) => {
        return ok(folder)
    }).catch((tempDirErr: Error) => {
        return err(tempDirErr)
    });

    if (tempDirPathResult.isErr()) {
        return err(new Error("Failed to create temporary directory for 'uploadFiles' testing."))
    }
    return ok(tempDirPathResult.value)
}

function getFileServerContainerConfigSupplier(filesArtifactMountPoints: Map<FilesArtifactID, string>):
(ipAddr: string) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = (ipAddr:string): Result<ContainerConfig, Error> => {

        const usedPorts = new Map<string, PortSpec>()
        usedPorts.set(FILE_SERVER_PORT_ID, FILE_SERVER_PORT_SPEC)

        const containerConfig = new ContainerConfigBuilder(FILE_SERVER_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withFiles(filesArtifactMountPoints)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier
}

async function getFileContents(ipAddress: string, portNum: number, filename: string): Promise<Result<any, Error>> {
    let response;
    try {
        response = await axios(`http://${ipAddress}:${portNum}/${filename}`)
    } catch(error){
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