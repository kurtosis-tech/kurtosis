/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_registration_service

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
)

// TODO rename to healthcheck_service
type LifecycleService struct {
	bindings.UnimplementedLifecycleServiceServer

	shutdownChan chan exit_codes.ApiContainerExitCode
}

// TODO rename to healthcheck_service
func NewLifecycleService() *LifecycleService {
	return &LifecycleService{}
}

func (service *LifecycleService) Startup() {
	go func() {

	}
}

func RegisterSuite() {

}

/*
func (service LifecycleService) Shutdown(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	service.shutdownChan <- "shutdown received"
	return &emptypb.Empty{}, nil
}

 */
