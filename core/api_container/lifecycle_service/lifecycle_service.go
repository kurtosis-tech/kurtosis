/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package lifecycle_service

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LifecycleService struct {
	bindings.UnimplementedLifecycleServiceServer
}

func NewLifecycleService() *LifecycleService {
	return &LifecycleService{}
}

func (service LifecycleService) IsAvailable(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}



