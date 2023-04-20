package grpc_file_streaming

import (
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
)

// ServerStream is a wrapper around a GRPC ServerStream object to be able to send and receive payloads bypassing the
// 4MB limit set by GRPC.
type ServerStream[DataChunkMessageType any, ServerResponseType any] struct {
	grpcStream grpc.ServerStream
}

func NewServerStream[DataChunkMessageType any, ServerResponseType any](
	grpcStream grpc.ServerStream,
) *ServerStream[DataChunkMessageType, ServerResponseType] {
	return &ServerStream[DataChunkMessageType, ServerResponseType]{
		grpcStream: grpcStream,
	}
}

// SendData sends some content via streaming by splitting it into fixed-sized chunks and streaming then to the client.
func (serverStream *ServerStream[DataChunkMessageType, ServerResponseType]) SendData(
	contentNameForLogging string,
	contentToSend []byte,
	grpcMsgConstructor func(previousChunkHash string, contentChunk []byte) (*DataChunkMessageType, error),
) error {
	// Split the content into chunks and stream them to the client
	err := sendMessagesToStream[DataChunkMessageType](
		contentNameForLogging,
		contentToSend,
		serverStream.grpcStream.SendMsg,
		grpcMsgConstructor,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sending '%s'", contentNameForLogging)
	}
	return nil
}

// ReceiveData receives some content via streaming expecting the content to be received in fixed-sized chunks from the
// client until the client returns an EOF. Once that happens, the chunks are assembled in passed to the
// assembledContentConsumer consumer function.
func (serverStream *ServerStream[DataChunkMessageType, ServerResponseType]) ReceiveData(
	contentNameForLogging string,
	grpcMsgExtractor func(dataChunk *DataChunkMessageType) ([]byte, string, error),
	assembledContentConsumer func(assembledContent []byte) (*ServerResponseType, error),
) error {
	// Read all the chunks and assemble them into a single byte array assembledContent
	assembledContent, err := readMessagesFromStream[DataChunkMessageType](
		contentNameForLogging,
		serverStream.grpcStream.RecvMsg,
		grpcMsgExtractor)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred sending '%s' through the stream", contentNameForLogging)
	}

	// Consume the fully assembled content and send the final response object through the stream if successful
	response, err := assembledContentConsumer(assembledContent)
	if err != nil {
		return stacktrace.Propagate(err, "Error consuming the assembled content for '%s' after all chunks were "+
			"received", contentNameForLogging)
	}

	// And send the final server response object corresponding, effectively closing the stream
	if err = serverStream.grpcStream.SendMsg(response); err != nil {
		return stacktrace.Propagate(err, "Error sending the final response object for '%s' through the stream",
			contentNameForLogging)
	}
	return nil
}
