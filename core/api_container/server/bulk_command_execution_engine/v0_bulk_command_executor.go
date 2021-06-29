/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package bulk_command_execution_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/palantir/stacktrace"
)

type v0BulkCommandExecutor struct {
	serviceNetworkProxy *server.ServiceNetworkRpcApiProxy
}

func (executor v0BulkCommandExecutor) VisitRegisterServiceCommand(args *core_api_bindings.RegisterServiceArgs) error {
	resp, err := executor.serviceNetworkProxy.RegisterService(args)
	if err != nil {
		return stacktrace.Propagate(err, "The service network proxy threw an error")
	}
	resp.IpAddr

	return nil
}

func (executor v0BulkCommandExecutor) VisitStartServiceCommand(args *core_api_bindings.StartServiceArgs) error {
	if _, err := executor.serviceNetworkProxy.StartService(context.Background(), args); err != nil {
		return stacktrace.Propagate(err, "The service network proxy threw an error")
	}
	return nil
}

func (executor v0BulkCommandExecutor) VisitRemoveServiceCommand(args *core_api_bindings.RemoveServiceArgs) error {
	panic("implement me")
}

