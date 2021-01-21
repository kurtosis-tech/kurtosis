/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_registration_service

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SuiteRegistrationService struct {
	// The action the newly-registered suite should take
	suiteAction bindings.SuiteAction

	isSuiteRegistered *server.ConcurrentBool
}

func NewSuiteRegistrationService(suiteAction bindings.SuiteAction, isSuiteRegistered *server.ConcurrentBool) *SuiteRegistrationService {
	return &SuiteRegistrationService{suiteAction: suiteAction, isSuiteRegistered: isSuiteRegistered}
}

func (s SuiteRegistrationService) RegisterSuite(ctx context.Context, empty *emptypb.Empty) (*bindings.SuiteRegistrationResponse, error) {
	response := &bindings.SuiteRegistrationResponse{SuiteAction: s.suiteAction}
	s.isSuiteRegistered.Set(true)
	return response, nil
}