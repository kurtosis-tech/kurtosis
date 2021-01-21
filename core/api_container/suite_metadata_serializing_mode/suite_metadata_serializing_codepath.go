/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_metadata_serializing_mode

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/server/suite_metadata_serialization"
	"path"
)

const (
	listenProtocol = "tcp"
)

type SuiteMetadataSerializingCodepath struct {
	listenAddress string
	args          SuiteMetadataSerializingArgs
}

func NewSuiteMetadataSerializingCodepath(args SuiteMetadataSerializingArgs) *SuiteMetadataSerializingCodepath {
	return &SuiteMetadataSerializingCodepath{args: args}
}

func (codepath SuiteMetadataSerializingCodepath) Execute() (int, error) {
	args := codepath.args


	serializedSuiteMetadataOutputFilepath := path.Join(
		api_container_docker_consts.SuiteExecutionVolumeMountDirpath,
		args.SuiteMetadataRelativeFilepath)
	suiteMetadataSerializingService := suite_metadata_serialization.newSuiteMetadataSerializingService(
		serializedSuiteMetadataOutputFilepath)
	bindings.RegisterSuiteMetadataSerializingServiceServer(codepath.grpcServer, suiteMetadataSerializingService)

	codepath.grpcServer


}

