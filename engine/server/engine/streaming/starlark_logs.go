package streaming

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
)

type AsyncStarlarkLogs struct {
	cancelCtxFunc               context.CancelFunc
	starlarkRunResponseLineChan chan *rpc_api.StarlarkRunResponseLine
}

func NewAsyncStarlarkLogs(cancelCtxFunc context.CancelFunc) AsyncStarlarkLogs {
	starlarkResponseLineChan := make(chan *rpc_api.StarlarkRunResponseLine)
	return AsyncStarlarkLogs{
		cancelCtxFunc:               cancelCtxFunc,
		starlarkRunResponseLineChan: starlarkResponseLineChan,
	}
}

func (async AsyncStarlarkLogs) Close() {
	logrus.Debugf("Streaming of Starlark execution logs is done, cleaning up resources")
	close(async.starlarkRunResponseLineChan)
	async.cancelCtxFunc()
}

func (async AsyncStarlarkLogs) AttachStream(stream grpc.ClientStream) {
	logrus.Debugf("Asynchronously reading the stream of Starlark execution logs")
	defer func() {
		async.Close()
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
		async.starlarkRunResponseLineChan <- responseLine
	}
}

func (async AsyncStarlarkLogs) WaitAndConsumeAll() []rpc_api.StarlarkRunResponseLine {
	var logs []*rpc_api.StarlarkRunResponseLine
	async.Consume(func(elem *rpc_api.StarlarkRunResponseLine) {
		if elem != nil {
			logs = append(logs, elem)
		}
	})
	return utils.FilterListNils(logs)
}

func (async AsyncStarlarkLogs) Consume(consumer func(*rpc_api.StarlarkRunResponseLine)) {
	for elem := range async.starlarkRunResponseLineChan {
		consumer(elem)
	}
}
