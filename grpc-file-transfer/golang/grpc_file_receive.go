package grpc_file_transfer

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"io"
)

func ReadBytesStream[DataChunkType interface{}, ResponseType interface{}](
	stream grpc.ServerStream,
	chunkExtractor func(dataChunk *DataChunkType) ([]byte, string, error),
	contentConsumer func(fullContent []byte) (*ResponseType, error),
) error {
	var content []byte
	var previousChunkHash string

	dataChunk := new(DataChunkType)
	errorReceivingChunk := stream.RecvMsg(dataChunk)
	for errorReceivingChunk == nil {
		chunkContent, previousChunkHashFromChunk, err := chunkExtractor(dataChunk)
		if err != nil {
			return stacktrace.NewError("An unexpected error occurred receiving a data chunk")
		}

		if previousChunkHashFromChunk != previousChunkHash {
			return stacktrace.NewError("An unexpected error occurred receiving file artifacts chunk. Hash validation did not pass")
		}
		content = append(content, chunkContent...)

		hasher := sha1.New()
		hasher.Write(chunkContent)
		previousChunkHash = hex.EncodeToString(hasher.Sum(nil))

		dataChunk = new(DataChunkType)
		errorReceivingChunk = stream.RecvMsg(dataChunk)
	}
	if errorReceivingChunk != io.EOF {
		return stacktrace.Propagate(errorReceivingChunk, "Unexpected error occurred receiving file artifact chunk")
	}

	response, err := contentConsumer(content)
	if err != nil {
		return stacktrace.Propagate(err, "Error consuming the entire content once received")
	}

	if err = stream.SendMsg(response); err != nil {
		return stacktrace.Propagate(err, "Error sending result through the stream")
	}
	return nil
}
