import {err, ok, Result, Err} from "neverthrow";
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";

const ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN = "upload-test-"
const ARCHIVE_SUBDIRECTORY_TEST_PATTERN = "sub-folder-"
const ARCHIVE_TEST_FILE_PATTERN = "test-file-"
const ARCHIVE_TEST_FILE_EXTENSION = ".txt"
const ARCHIVE_TEST_CONTENT = "This is file is for testing purposes."

const NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY = 3
const NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY = 1

const ENCLAVE_TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = true

test("Test Upload Files", async () => {
    //Make directory for usage.
    const tempDirectory = await createTestFolderToUpload()
    if (tempDirectory.isErr()) { throw tempDirectory.error }
})

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