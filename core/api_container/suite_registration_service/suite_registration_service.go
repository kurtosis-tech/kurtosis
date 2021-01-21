/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_registration_service

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SuiteRegistrationService struct {
	// The action the newly-registered suite should take
	suiteAction bindings.SuiteAction

	suiteRegistrationChan chan interface{}
}

func NewSuiteRegistrationService(suiteAction bindings.SuiteAction, suiteRegistrationChan chan interface{}) *SuiteRegistrationService {
	return &SuiteRegistrationService{suiteAction: suiteAction, suiteRegistrationChan: suiteRegistrationChan}
}

func (s SuiteRegistrationService) RegisterSuite(_ context.Context, _ *emptypb.Empty) (*bindings.SuiteRegistrationResponse, error) {
	response := &bindings.SuiteRegistrationResponse{SuiteAction: s.suiteAction}
	s.suiteRegistrationChan <- "Suite registered"
	return response, nil
}