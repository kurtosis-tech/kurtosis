import {
    ContainerConfig,
    ContainerConfigBuilder,
    FilesArtifactUUID,
    PortProtocol,
    PortSpec,
    ServiceID
} from "kurtosis-sdk";
import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";
import {err, ok, Result} from "neverthrow";
import axios from "axios";

const TEST_NAME = "duplicate-files-artifact-mount"
const IS_PARTITIONING_ENABLED = false

const IMAGE = "flashspys/nginx-static"
const SERVICE_ID: ServiceID = "file-server"

const TEST_FILES_ARTIFACT_URL = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

const USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT  = "/static"

const DUPLICATE_MOUNTPOINT_DOCKER_DAEMON_ERR_MSG  = "Duplicate mount point"

const DUPLICATE_MOUNTPOINT_KUBERNETES_ERR_MSG = "mountPath: Invalid value: \"/static\": must be unique"

jest.setTimeout(180000)

test("Test two files artifacts mounted to the same location", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {

        // ------------------------------------- TEST SETUP ----------------------------------------------
        const storeFirstArtifactResult = await enclaveContext.storeWebFiles(TEST_FILES_ARTIFACT_URL)
        if (storeFirstArtifactResult.isErr()) { throw storeFirstArtifactResult.error }
        const firstFilesArtifactUuid = storeFirstArtifactResult.value;
        const storeSecondArtifactResult = await enclaveContext.storeWebFiles(TEST_FILES_ARTIFACT_URL)
        if (storeSecondArtifactResult.isErr()) { throw storeSecondArtifactResult.error }
        const secondFilesArtifactUuid = storeSecondArtifactResult.value;

        const filesArtifactMountpoints = new Map<FilesArtifactUUID, string>()
        filesArtifactMountpoints.set(firstFilesArtifactUuid, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT);
        filesArtifactMountpoints.set(secondFilesArtifactUuid, USER_SERVICE_MOUNTPOINT_FOR_TEST_FILESARTIFACT);
        const fileServerContainerConfig = getFileServerContainerConfig(filesArtifactMountpoints)


        // ------------------------------------- TEST RUN ----------------------------------------------
        const addServiceResult = await enclaveContext.addService(SERVICE_ID, fileServerContainerConfig)
        if(addServiceResult.isOk()){
            throw new Error(`Adding service '${SERVICE_ID}' should have failed and did not, because duplicated files artifact mountpoints ` +
                `'${filesArtifactMountpoints}' should throw an error`)
        }
        const errMsg = addServiceResult.error.message
        if(!errMsg.includes(DUPLICATE_MOUNTPOINT_DOCKER_DAEMON_ERR_MSG) && !errMsg.includes(DUPLICATE_MOUNTPOINT_KUBERNETES_ERR_MSG)){
            throw new Error(`Adding service "${SERVICE_ID}" has failed, but the error is not the duplicated-files-artifact-mountpoints-error that we expected, this is throwing this error instead:\n "${errMsg}"`)
        }
    }finally{
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

function getFileServerContainerConfig(filesArtifactMountpoints: Map<FilesArtifactUUID, string>): ContainerConfig {
    const containerConfig = new ContainerConfigBuilder(IMAGE)
        .withFiles(filesArtifactMountpoints)
        .build()

   return containerConfig
}