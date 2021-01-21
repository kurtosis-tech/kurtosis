/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/execution_codepath"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/suite_metadata_serialization"
	"github.com/kurtosis-tech/kurtosis/api_container/suite_metadata_serializing_mode"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

func Create(mode api_container_env_vars.ApiContainerMode, paramsJson string) (server.ApiContainerServerCore, error) {
	paramsJsonBytes := []byte(paramsJson)

	var result server.ApiContainerServerCore
	switch mode {
	case api_container_env_vars.SuiteMetadataSerializingMode:
		var args suite_metadata_serializing_mode.SuiteMetadataSerializingArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the suite metadata serializing args JSON")
		}
		serializationOutputFilepath := path.Join(
			api_container_docker_consts.SuiteExecutionVolumeMountDirpath,
			args.SuiteMetadataRelativeFilepath)
		result = suite_metadata_serialization.NewSuiteMetadataSerializingServerCore(serializationOutputFilepath)
	case api_container_env_vars.TestExecutionMode:
		var args test_execution_mode.TestExecutionArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the test execution args JSON")
		}
	default:
		return nil, stacktrace.NewError("Unrecognized API container mode '%v'", mode)
	}
}
