import {createEnclave} from "../../test_helpers/enclave_setup";
import {DEFAULT_DRY_RUN, EMPTY_RUN_PARAMS, IS_PARTITIONING_ENABLED, JEST_TIMEOUT_MS} from "./shared_constants";
import * as path from "path";
import log from "loglevel";
import {err} from "neverthrow";

const INVALID_KURTOSIS_YAML_TEST_NAME = "invalid-package-invalid-yaml-file"
const INVALID_KURTOSIS_YAML_IN_PACKAGE_REL_PATH = "../../../../starlark/invalid-yaml-file"

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test invalid package with invalid yaml file", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(INVALID_KURTOSIS_YAML_TEST_NAME, IS_PARTITIONING_ENABLED)

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const packageRootPath = path.join(__dirname, INVALID_KURTOSIS_YAML_IN_PACKAGE_REL_PATH)

        log.info(`Loading package at path '${packageRootPath}'`)

        const runResult = await enclaveContext.runStarlarkPackageBlocking(packageRootPath, EMPTY_RUN_PARAMS, DEFAULT_DRY_RUN)

        if (!runResult.isErr()) {
            throw err(new Error("Package with invalid package was expected to error but didn't"))
        }

        if (!runResult.error.message.includes(`Field 'name', which is the Starlark package's name, in 'kurtosis.yml' needs to be set and cannot be empty`)) {
            throw err(new Error(`Unexpected error message. The received error is:\n${runResult.error.message}`))
        }
    } finally {
        stopEnclaveFunction()
    }
})
