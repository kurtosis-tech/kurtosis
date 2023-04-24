package grpc_file_streaming

import (
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
)

// ClientStream is a wrapper around a GRPC ClientStream object to be able to send and receive payloads bypassing the
// 4MB limit set by GRPC.
type ClientStream[DataChunkMessageType any, ServerResponseType any] struct {
	grpcStream grpc.ClientStream
}

func NewClientStream[DataChunkMessageType any, ServerResponseType any](
	grpcStream grpc.ClientStream,
) *ClientStream[DataChunkMessageType, ServerResponseType] {
	return &ClientStream[DataChunkMessageType, ServerResponseType]{
		grpcStream: grpcStream,
	}
}

// SendData sends some content via streaming by splitting it into fixed-sized chunks and streaming them to the server.
// Once all chunks are sent, it calls CloseSend() on the GRPC underlying stream and expected the server to send
// back the final response object.
func (clientStream *ClientStream[DataChunkMessageType, ServerResponseType]) SendData(
	contentNameForLogging string,
	contentToSend []byte,
	grpcMsgConstructor func(previousChunkHash string, contentChunk []byte) (*DataChunkMessageType, error),
) (*ServerResponseType, error) {
	// Split the content into chunks and stream them to the server
	err := sendMessagesToStream[DataChunkMessageType](
		contentNameForLogging,
		contentToSend,
		clientStream.grpcStream.SendMsg,
		grpcMsgConstructor,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred sending '%s'", contentNameForLogging)
	}

	// The client (which is the sender here) needs to close the stream to tell the server nothing more is expected
	if err = clientStream.grpcStream.CloseSend(); err != nil {
		return nil, stacktrace.Propagate(err, "An error was encountered closing the stream for '%s' after all "+
			"chunks were sent", contentNameForLogging)
	}

	// In response to the close, the server sends the response type through the stream before completely closing it
	response := new(ServerResponseType)
	if err = clientStream.grpcStream.RecvMsg(response); err != nil {
		return nil, stacktrace.Propagate(err, "Error receiving the final response object for '%s' through the stream",
			contentNameForLogging)
	}
	return response, nil
}

// ReceiveData receives some content via streaming expecting the content to be received in fixed-sized chunks from
// the server until the server returns io.EOF. Once that happens, the chunks are assembled into a single byte array
// and returned.
func (clientStream *ClientStream[DataChunkMessageType, ServerResponseType]) ReceiveData(
	contentNameForLogging string,
	grpcMsgExtractor func(dataChunk *DataChunkMessageType) ([]byte, string, error),
) ([]byte, error) {
	// Read all the chunks and assemble them into a single byte array assembledContent
	assembledContent, err := readMessagesFromStream[DataChunkMessageType](
		contentNameForLogging,
		clientStream.grpcStream.RecvMsg,
		grpcMsgExtractor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred sending the data chunks for '%s' through the stream",
			contentNameForLogging)
	}
	return assembledContent, nil
}
