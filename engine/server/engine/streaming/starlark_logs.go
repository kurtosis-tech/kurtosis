package streaming

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	rpc_api "github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/stacktrace"
)

type asyncStarlarkLogs struct {
	ctx                         context.Context
	cancelCtxFunc               context.CancelFunc
	starlarkRunResponseLineChan chan *rpc_api.StarlarkRunResponseLine
	markedForConsumption        bool
}

func NewAsyncStarlarkLogs(cancelCtxFunc context.CancelFunc) asyncStarlarkLogs {
	ctx, cancel := context.WithCancel(context.Background())
	starlarkResponseLineChan := make(chan *rpc_api.StarlarkRunResponseLine)
	markedForConsumption := false
	return asyncStarlarkLogs{
		ctx:                         ctx,
		cancelCtxFunc:               cancel,
		starlarkRunResponseLineChan: starlarkResponseLineChan,
		markedForConsumption:        markedForConsumption,
	}
}

func (async *asyncStarlarkLogs) MarkForConsumption() {
	markedForConsumption := true
	async.markedForConsumption = markedForConsumption
}

func (async *asyncStarlarkLogs) IsMarkedForConsumption() bool {
	return async.markedForConsumption
}

func (async *asyncStarlarkLogs) Close() {
	logrus.Debugf("Streaming of Starlark execution logs is done, cleaning up resources")
	async.cancelCtxFunc()
}

func (async *asyncStarlarkLogs) AttachStream(stream grpc.ClientStream) {
	logrus.Debugf("Asynchronously reading the stream of Starlark execution logs")
	defer func() {
		close(async.starlarkRunResponseLineChan)
	}()
	for {
		responseLine := new(rpc_api.StarlarkRunResponseLine)
		err := stream.RecvMsg(responseLine)
		if err == io.EOF {
			logrus.Debugf("Successfully reached the end of the response stream. Closing.")
			return
		}
		if err != nil {
			logrus.Errorf("Unexpected error happened reading the stream. Client might have cancelled the stream\n%v", err.Error())
			return
		}
		select {
		case <-async.ctx.Done():
			logrus.Debugf("Resources have been closed before consuming all the upstream data")
			return
		case async.starlarkRunResponseLineChan <- responseLine:
		}
	}
}

func (async *asyncStarlarkLogs) WaitAndConsumeAll() ([]rpc_api.StarlarkRunResponseLine, error) {
	var logs []*rpc_api.StarlarkRunResponseLine
	err := async.Consume(func(elem *rpc_api.StarlarkRunResponseLine) error {
		if elem != nil {
			logs = append(logs, elem)
		}
		return nil
	})
	notNilsLogs := utils.FilterListNils(logs)

	if err != nil {
		return notNilsLogs, stacktrace.Propagate(err, "Failed to consume all logs, %d were consumed before the error", len(notNilsLogs))
	}

	return notNilsLogs, nil
}

func (async *asyncStarlarkLogs) Consume(consumer func(*rpc_api.StarlarkRunResponseLine) error) error {
	for elem := range async.starlarkRunResponseLineChan {
		if err := consumer(elem); err != nil {
			return stacktrace.Propagate(err, "Failed to consume element of type '%T'", elem)
		}
	}
	return nil
}
