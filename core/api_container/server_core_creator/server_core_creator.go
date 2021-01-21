/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/suite_metadata_serialization"
	"github.com/palantir/stacktrace"
	"path"
)

func Create(mode api_container_env_vars.ApiContainerMode, paramsJson string) (server.ApiContainerServerCore, error) {
	paramsJsonBytes := []byte(paramsJson)

	switch mode {
	case api_container_env_vars.SuiteMetadataSerializingMode:
		var args SuiteMetadataSerializingArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the suite metadata serializing args JSON")
		}
		serializationOutputFilepath := path.Join(
			api_container_docker_consts.SuiteExecutionVolumeMountDirpath,
			args.SuiteMetadataRelativeFilepath)
		result := suite_metadata_serialization.NewSuiteMetadataSerializationServerCore(serializationOutputFilepath)
		return result,  nil
	case api_container_env_vars.TestExecutionMode:
		var args TestExecutionArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the test execution args JSON")
		}
		return result, nil
	default:
		return nil, stacktrace.NewError("Unrecognized API container mode '%v'", mode)
	}
}
