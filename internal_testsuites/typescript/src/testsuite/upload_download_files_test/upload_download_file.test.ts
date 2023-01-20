import {err, ok, Result} from "neverthrow";
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {checkFileContents, startFileServer} from "../../test_helpers/test_helpers";
import {ServiceName} from "kurtosis-sdk";

const ARCHIVE_DIRECTORY_TEST_PATTERN   = "upload-test-typescript-"
const ARCHIVE_SUBDIRECTORY_TEST_PATTERN     = "sub-folder-"
const ARCHIVE_TEST_FILE_PATTERN             = "test-file-"
const ARCHIVE_TEST_FILE_EXTENSION           = ".txt"
const ARCHIVE_TEST_CONTENT                  = "This file is for testing purposes."
const TEST_ARTIFACT_NAME                    = "test-artifact"

const NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY      = 3
const NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY    = 1

const ENCLAVE_TEST_NAME         = "upload-files-test"
const IS_PARTITIONING_ENABLED   = false

//Keywords for mapping paths for file integrity checking.
const DISK_DIR_KEYWORD                                  = "diskDir"
const ARCHIVE_DIR_KEYWORD                               = "archiveDir"
const SUB_DIR_KEYWORD                                   = "subDir"
const SUB_FILE_KEYWORD_PATTERN                          = "subFile"
const ARCHIVE_ROOT_FILE_KEYWORD_PATTERN                 = "archiveRootFile"

const FOLDER_PERMISSION = 0o755
const FILE_PERMISSION   = 0o644

const FILE_SERVER_SERVICE_NAME : ServiceName = "file-server"

jest.setTimeout(180000)

test("Test Upload and Download Files", TestUploadAndDownloadFiles)

async function TestUploadAndDownloadFiles() {
    const testFolderResults = await createTestFolderToUpload()
    if (testFolderResults.isErr()) { throw testFolderResults.error }
    const filePathsMap = testFolderResults.value

    const createEnclaveResult = await createEnclave(ENCLAVE_TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value
    try {
        const pathToUpload = filePathsMap.get(DISK_DIR_KEYWORD)
        if (typeof pathToUpload === "undefined") {throw new Error("Failed to store uploadable path in path map.")}
        const uploadResults = await enclaveContext.uploadFiles(pathToUpload, TEST_ARTIFACT_NAME)
        if(uploadResults.isErr()) { throw uploadResults.error }
        const artifactUuid = uploadResults.value;

        const firstArchiveRootKeyWord = `${ARCHIVE_ROOT_FILE_KEYWORD_PATTERN}0`
        const firstArchiveRootFilename = `${filePathsMap.get(firstArchiveRootKeyWord)}`

        const startFileServerResult = await startFileServer(FILE_SERVER_SERVICE_NAME, TEST_ARTIFACT_NAME, firstArchiveRootFilename, enclaveContext)
        if (startFileServerResult.isErr()){throw startFileServerResult.error}
        const {fileServerPublicIp, fileServerPublicPortNum} = startFileServerResult.value

        const allContentTestResults = await testAllContent(filePathsMap, fileServerPublicIp, fileServerPublicPortNum)
        if(allContentTestResults.isErr()) { throw allContentTestResults.error}

        const downloadFilesViaUuidResult = await enclaveContext.downloadFilesArtifact(artifactUuid)
        if (downloadFilesViaUuidResult.isErr()) {throw downloadFilesViaUuidResult.error}
        const downloadFilesViaShortenedUuidResult = await enclaveContext.downloadFilesArtifact(artifactUuid.substring(0, 12))
        if (downloadFilesViaShortenedUuidResult.isErr()) {throw downloadFilesViaShortenedUuidResult.error}
        const downloadFilesViaNameResult = await enclaveContext.downloadFilesArtifact(TEST_ARTIFACT_NAME)
        if (downloadFilesViaNameResult.isErr()) {throw downloadFilesViaNameResult.error}

        for (let i = 0;  i < downloadFilesViaNameResult.value.length; i++) {
            if (downloadFilesViaUuidResult.value.at(i) != downloadFilesViaShortenedUuidResult.value.at(i) || downloadFilesViaUuidResult.value.at(i) != downloadFilesViaNameResult.value.at(i)) {
                throw new Error("Expected files downloaded via uuid, shortened uuid and name to be the same but weren't")
            }
        }

    } finally {
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
}

//========================================================================
// Helpers
//========================================================================
async function testAllContent(
    allPaths: Map<string,string>,
    ipAddress: string,
    portNum: number
): Promise<Result<null, Error>>{
    //Test files in archive root directory.
    const rootDirTestResults = await testDirectoryContents(
        allPaths,
        ARCHIVE_ROOT_FILE_KEYWORD_PATTERN,
        NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY,
        ipAddress,
        portNum
    )
    if(rootDirTestResults.isErr()){ return err(rootDirTestResults.error) }

    //Test files in subdirectory.
    const subDirTestResults = await testDirectoryContents(
        allPaths,
        SUB_FILE_KEYWORD_PATTERN,
        NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY,
        ipAddress,
        portNum
    )
    if (subDirTestResults.isErr()){ return err(subDirTestResults.error) }

    return ok(null)
}

//Cycle through a directory and check the file contents.
async function testDirectoryContents(
    pathsMap: Map<string, string>,
    fileKeywordPattern: string,
    fileCount: number,
    ipAddress: string,
    portNum: number
): Promise<Result<null, Error>> {

    for(let i = 0; i < fileCount; i++){
        let fileKeyword = `${fileKeywordPattern}${i}`
        let relativePath = pathsMap.get(fileKeyword)
        if (typeof relativePath === "undefined"){
            return err(new Error(`The file for keyword ${fileKeyword} was not mapped in the paths map.`))
        }
        let testContentResults = await checkFileContents(ipAddress, portNum, relativePath, ARCHIVE_TEST_CONTENT)
        if (testContentResults.isErr()) { return  err(testContentResults.error) }
    }
    return ok(null)
}

async function createTestFiles(pathToCreateAt : string, fileCount : number): Promise<Result<string[], Error>>{
    let filenames: string[] = []

    for (let i = 0; i < fileCount; i++) {
        const filename = `${ARCHIVE_TEST_FILE_PATTERN}${i}${ARCHIVE_TEST_FILE_EXTENSION}`
        const fullFilepath = path.join(pathToCreateAt, filename)
        filenames.push(filename)
        try {
            await filesystem.promises.writeFile(fullFilepath, ARCHIVE_TEST_CONTENT)
            await filesystem.promises.chmod(fullFilepath, FILE_PERMISSION)
        } catch {
            return err(new Error(`Failed to write test file '${filename}' at '${fullFilepath}'.`))
        }
    }
    return ok(filenames)
}

//Creates a temporary folder with x files and 1 sub folder that has y files each.
//Where x is numberOfTempTestFilesToCreateInArchiveDir
//Where y is numberOfTempTestFilesToCreateInSubDir
async function createTestFolderToUpload(): Promise<Result<Map<string,string>, Error>> {
    const testDirectory = os.tmpdir()

    //Create a base directory
    const tempDirectoryResult = await createTempDirectory(testDirectory, ARCHIVE_DIRECTORY_TEST_PATTERN)
    if(tempDirectoryResult.isErr()) { return err(tempDirectoryResult.error) }
    const baseTempDirPath = tempDirectoryResult.value

    //Create a single subdirectory.
    const subDirectoryResult = await  createTempDirectory(baseTempDirPath, ARCHIVE_SUBDIRECTORY_TEST_PATTERN)
    if(subDirectoryResult.isErr()) { return err(subDirectoryResult.error) }
    const tempSubDirectory = subDirectoryResult.value

    //Create NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY
    const subDirTestFileResult = await createTestFiles(tempSubDirectory, NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY)
    if(subDirTestFileResult.isErr()){ return err(subDirTestFileResult.error)}

    //Create NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY
    const rootDirTestFileResults = await createTestFiles(baseTempDirPath, NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY)
    if(rootDirTestFileResults.isErr()) { return err(rootDirTestFileResults.error) }

    //Set folder permissions.
    try {
        await filesystem.promises.chmod(tempDirectoryResult.value, FOLDER_PERMISSION) //baseDirectory
    } catch {
        return err(new Error(`Failed to set permissions for '${tempDirectoryResult.value}'`))
    }
    try {
        await filesystem.promises.chmod(subDirectoryResult.value, FOLDER_PERMISSION) //subdirectory
    } catch {
        return err(new Error(`Failed to set permissions for '${subDirectoryResult.value}'`))
    }

    let subDirFilenames = subDirTestFileResult.value
    let archiveRootFilenames = rootDirTestFileResults.value
    let archiveRootDir = path.basename(baseTempDirPath)
    let subDir = path.basename(tempSubDirectory)

    let relativeDiskPaths = new Map<string, string>()
    relativeDiskPaths.set(DISK_DIR_KEYWORD, `${testDirectory}/${archiveRootDir}`)
    relativeDiskPaths.set(ARCHIVE_DIR_KEYWORD, archiveRootDir)
    relativeDiskPaths.set(SUB_DIR_KEYWORD, `${archiveRootDir}/${subDir}`)

    for(let i = 0; i < subDirFilenames.length; i++){
        let keyword = `${SUB_FILE_KEYWORD_PATTERN}${i}`
        let basename = path.basename(subDirFilenames[i])
        let relativeSubFile = `${subDir}/${basename}`
        relativeDiskPaths.set(keyword, relativeSubFile)
    }

    for(let i = 0; i < archiveRootFilenames.length; i++){
        let keyword = `${ARCHIVE_ROOT_FILE_KEYWORD_PATTERN}${i}`
        let basename = path.basename(archiveRootFilenames[i])
        relativeDiskPaths.set(keyword, basename)
    }
    return ok(relativeDiskPaths)
}

//Stand in for go's ioutil.TempDir
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
