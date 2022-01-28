import log from "loglevel";

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "bulk-command-execution-test";
const IS_PARTITIONING_ENABLED = true;
const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started";

jest.setTimeout(180000)

test("Test bulk command execution", async () => {

    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST RUN ----------------------------------------------
        log.info("Executing JSON-serialized commands to create a network with various services and repartition it...")

        const bulkCommandJson = generateBulkCommandJson()

        const executeBulkCommandsResult = await enclaveContext.executeBulkCommands(bulkCommandJson)

        if(executeBulkCommandsResult.isErr()) { throw executeBulkCommandsResult.error }

        log.info("Successfully executed JSON-serialized commands")

    }finally{
        stopEnclaveFunction()
    }
})

function generateBulkCommandJson() {
    const result = `{
        "schemaVersion": 0,
        "body": {
            "commands": [
                {
                    "type": "REGISTER_SERVICE",
                    "args": {
                        "service_id": "service1"
                    }
                },
                {
                    "type": "START_SERVICE",
                    "args": {
                        "service_id": "service1",
                        "docker_image": "${DOCKER_GETTING_STARTED_IMAGE}",
                        "used_ports": {
                            "80/tcp": true
                        },
                        "enclave_data_dir_mnt_dirpath": "/kurtosis-enclave-data"
                    }
                },
                {
                    "type": "WAIT_FOR_HTTP_GET_ENDPOINT_AVAILABILITY",
                    "args": {
                        "service_id": "service1",
                        "port": 80,
                        "path": "",
                        "initial_delay_milliseconds": 0,
                        "retries": 5,
                        "retries_delay_milliseconds": 2000,
                        "body_text": ""
                    }
                },
                {
                    "type": "REGISTER_SERVICE",
                    "args": {
                        "service_id": "service2"
                    }
                },
                {
                    "type": "START_SERVICE",
                    "args": {
                        "service_id": "service2",
                        "docker_image": "${DOCKER_GETTING_STARTED_IMAGE}",
                        "used_ports": {
                            "80/tcp": true
                        },
                        "enclave_data_dir_mnt_dirpath": "/kurtosis-enclave-data"
                    }
                },
                {
                    "type": "WAIT_FOR_HTTP_GET_ENDPOINT_AVAILABILITY",
                    "args": {
                        "service_id": "service2",
                        "port": 80,
                        "path": "",
                        "initial_delay_milliseconds": 0,
                        "retries": 5,
                        "retries_delay_milliseconds": 2000,
                        "body_text": ""
                    }
                },
                {
                    "type": "REPARTITION",
                    "args": {
                        "partition_services": {
                            "partition1": {
                                "service_id_set": {
                                    "service1": true
                                }
                            },
                            "partition2": {
                                "service_id_set": {
                                    "service2": true
                                }
                            }
                        },
                        "default_connection": {
                            "is_blocked": true
                        }
                    }
                },
                {
                    "type": "REGISTER_SERVICE",
                    "args": {
                        "service_id": "service3",
                        "partition_id": "partition2"
                    }
                },
                {
                    "type": "START_SERVICE",
                    "args": {
                        "service_id": "service3",
                        "docker_image": "${DOCKER_GETTING_STARTED_IMAGE}",
                        "used_ports": {
                            "80/tcp": true
                        },
                        "enclave_data_dir_mnt_dirpath": "/kurtosis-enclave-data"
                    }
                },
                {
                    "type": "WAIT_FOR_HTTP_GET_ENDPOINT_AVAILABILITY",
                    "args": {
                        "service_id": "service3",
                        "port": 80,
                        "path": "",
                        "initial_delay_milliseconds": 0,
                        "retries": 5,
                        "retries_delay_milliseconds": 2000,
                        "body_text": ""
                    }
                },
                {
                    "type": "REPARTITION",
                    "args": {
                        "partition_services": {
                            "partition1": {
                                "service_id_set": {
                                    "service1": true,
                                    "service2": true,
                                    "service3": true
                                }
                            }
                        },
                        "default_connection": {
                            "is_blocked": false
                        }
                    }
                }
            ]
        }
    }`

    return result;
};