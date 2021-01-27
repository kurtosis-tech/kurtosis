/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_metadata_serialization

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"google.golang.org/grpc"
)

type SuiteMetadataSerializationServerCore struct {
	serializationOutputFilepath string
}

func NewSuiteMetadataSerializationServerCore(serializationOutputFilepath string) *SuiteMetadataSerializationServerCore {
	return &SuiteMetadataSerializationServerCore{serializationOutputFilepath: serializationOutputFilepath}
}

func (core SuiteMetadataSerializationServerCore) GetSuiteAction() bindings.SuiteAction {
	return bindings.SuiteAction_SERIALIZE_SUITE_METADATA
}

func (core SuiteMetadataSerializationServerCore) CreateAndRegisterService(
		shutdownChan chan int,
		grpcServer *grpc.Server) server.ApiContainerServerService {
	service := newSuiteMetadataSerializationService(core.serializationOutputFilepath, shutdownChan)
	bindings.RegisterSuiteMetadataSerializationServiceServer(grpcServer, service)
	return service
}



