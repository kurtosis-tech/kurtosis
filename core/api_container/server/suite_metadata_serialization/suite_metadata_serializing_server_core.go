/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_metadata_serialization

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"google.golang.org/grpc"
)

type SuiteMetadataSerializingServerCore struct {
	serializationOutputFilepath string
}

func NewSuiteMetadataSerializingServerCore(serializationOutputFilepath string) *SuiteMetadataSerializingServerCore {
	return &SuiteMetadataSerializingServerCore{serializationOutputFilepath: serializationOutputFilepath}
}

func (core SuiteMetadataSerializingServerCore) GetSuiteAction() bindings.SuiteAction {
	return bindings.SuiteAction_SERIALIZE_SUITE_METADATA
}

func (core SuiteMetadataSerializingServerCore) CreateAndRegisterService(shutdownChan chan exit_codes.ApiContainerExitCode, grpcServer *grpc.Server) server.ApiContainerServerService {
	service := newSuiteMetadataSerializingService(core.serializationOutputFilepath, shutdownChan)
	bindings.RegisterSuiteMetadataSerializingServiceServer(grpcServer, service)
	return service
}



