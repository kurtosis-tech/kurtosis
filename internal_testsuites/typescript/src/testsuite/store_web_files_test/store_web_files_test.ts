import { ServiceName } from "kurtosis-sdk"
import log from "loglevel";
import { Result, ok, err } from "neverthrow";
import axios from "axios"

import { createEnclave } from "../../test_helpers/enclave_setup";
import {startFileServer} from "../../test_helpers/test_helpers";

const TEST_NAME = "files-artifact-mounting"
const IS_PARTITIONING_ENABLED = false

const FILE_SERVER_SERVICE_NAME: ServiceName = "file-server"

const TEST_FILES_ARTIFACT_URL = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"
const TEST_ARTIFACT_NAME = "test-artifact"


// Filenames & contents for the files stored in the files artifact
const FILE1_FILENAME = "file1.txt"
const FILE2_FILENAME = "file2.txt"

const EXPECTED_FILE1_CONTENTS = "file1\n"
const EXPECTED_FILE2_CONTENTS = "file2\n"

jest.setTimeout(180000)

test.skip("Test web file storing", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {

        // ------------------------------------- TEST SETUP ----------------------------------------------
        const storeWebFilesResult = await enclaveContext.storeWebFiles(TEST_FILES_ARTIFACT_URL, TEST_ARTIFACT_NAME);
        if(storeWebFilesResult.isErr()) { throw storeWebFilesResult.error }
        const filesArtifactUuid = storeWebFilesResult.value;

        const startFileServerResult = await startFileServer(FILE_SERVER_SERVICE_NAME, filesArtifactUuid, FILE1_FILENAME, enclaveContext)
        if(startFileServerResult.isErr()) { throw startFileServerResult.error }
        const {fileServerPublicIp, fileServerPublicPortNum} = startFileServerResult.value
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
    }finally{
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

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