/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package bulk_command_execution_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	dockerGettingStartedImage = "docker/getting-started"
)

type BulkCommandExecutionTest struct {
}

func NewBulkCommandExecutionTest() *BulkCommandExecutionTest {
	return &BulkCommandExecutionTest{}
}

func (test BulkCommandExecutionTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60).WithPartitioningEnabled(true)
}

func (test BulkCommandExecutionTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	return networkCtx, nil
}

func (test BulkCommandExecutionTest) Run(network networks.Network) error {
	networkCtx, ok := network.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic network object")
	}

	logrus.Info("Executing JSON-serialized commands to create a network with various services and repartition it...")
	bulkCommandJson := generateBulkCommandJson()
	if err := networkCtx.ExecuteBulkCommands(bulkCommandJson); err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the bulk command JSON to set up the network")
	}
	logrus.Info("Successfully executed JSON-serialized commands")
	return nil
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
