package grpc_file_transfer

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"math"
)

const (
	chunkSize = 2 * 1024 * 1024 // 2mb
)

func SendBytesStream[DataChunkType interface{}, ResponseType interface{}](
	stream grpc.ClientStream,
	content []byte,
	chunkConsumer func(previousChunkHash string, chunkContext []byte) *DataChunkType,
) (*ResponseType, error) {
	var previousChunkHash string
	chunkNumber := 0
	totalChunkNumber := int(math.Ceil(float64(len(content)) / float64(chunkSize)))
	for idx := 0; idx < len(content); idx += chunkSize {
		chunkNumber += 1
		var contentChunk []byte
		if idx+chunkSize < len(content) {
			contentChunk = make([]byte, chunkSize)
			copy(contentChunk, content[idx:idx+chunkSize])
		} else {
			contentChunk = make([]byte, len(content)-idx)
			copy(contentChunk, content[idx:len(content)])
		}

		logrus.Debugf("Sending chunk number '%d'/'%d'", chunkNumber, totalChunkNumber)

		message := chunkConsumer(previousChunkHash, contentChunk)

		err := stream.SendMsg(message)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error was encountered while uploading data to the API Container.")
		}

		hasher := sha1.New()
		hasher.Write(contentChunk)
		previousChunkHash = hex.EncodeToString(hasher.Sum(nil))
		logrus.Debugf("Chunk number '%d/%d' sent. Hash was '%s'", chunkNumber, totalChunkNumber, previousChunkHash)
	}

	if err := stream.CloseSend(); err != nil {
		return nil, stacktrace.Propagate(err, "An error was encountered closing the byte stream for data transfer.")
	}

	response := new(ResponseType)
	if err := stream.RecvMsg(response); err != nil {
		return nil, stacktrace.Propagate(err, "An error was encountered receiving final response object from remote.")
	}
	return response, nil
}
