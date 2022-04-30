import {err, ok, Result, Err} from "neverthrow";
import * as filesystem from "fs"
import * as path from "path"
import * as os from "os";

const ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN = "upload-test-"
const ARCHIVE_SUBDIRECTORY_TEST_PATTERN = "sub-folder-"
const ARCHIVE_TEST_FILE_PATTERN = "test-file-"
const ARCHIVE_TEST_CONTENT = "This is file is for testing purposes."

const NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY = 3
const NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY = 1

const ENCLAVE_TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = true

test("Test Upload Files", async () => {

    //Make directory for usage.
    const osTempDirpath = os.tmpdir()
    const tempDirpathPrefix = path.join(osTempDirpath, ARCHIVE_ROOT_DIRECTORY_TEST_PATTERN)
    const tempDirpathResult = await filesystem.promises.mkdtemp(
        tempDirpathPrefix,
    ).then((folder: string) => {
        return ok(folder)
    }).catch((tempDirErr: Error) => {
        return err(tempDirErr)
    });
    if (tempDirpathResult.isErr()) {
        return err(new Error("Failed to create temporary directory for 'uploadFiles' testing.."))
    }
    const tempDirpath = tempDirpathResult.value

    throw new Err("This test is not implemented yet.")
})

//========================================================================
// Helpers
//========================================================================
async function createTestFiles(): Promise<Result<null, Error>>{
    return err(new Error("CreateTestFiles is not Implemented."))
}

async function createTestFolderToUpload(): Promise<Result<null, Error>> {
    return err(new Error("CreateTestFolderToUpload is not implemented."))
}