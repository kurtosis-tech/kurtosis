package bulk_command_execution_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "bulk-command-execution-test"
	isPartitioningEnabled = true

	dockerGettingStartedImage = "docker/getting-started"
)

func TestBulkCommandExecution(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing JSON-serialized commands to create a network with various services and repartition it...")
	bulkCommandJson := generateBulkCommandJson()
	require.NoError(t, enclaveCtx.ExecuteBulkCommands(bulkCommandJson), "An error occurred executing the bulk command JSON to set up the network")
	logrus.Info("Successfully executed JSON-serialized commands")
}

func generateBulkCommandJson() string {
	result := fmt.Sprintf(
		`
{
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
                    "docker_image": "%v",
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
                    "docker_image": "%v",
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
                    "docker_image": "%v",
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
}
`,
		dockerGettingStartedImage,
		dockerGettingStartedImage,
		dockerGettingStartedImage,
	)
	return result
}
