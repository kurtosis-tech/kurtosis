package grpc_file_streaming

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"math"
)

const (
	// Kurtosis imposes maximum file size of 100MB
	kurtosisMaximumFileSizeLimit = 100 * 1024 * 1024

	// Files are streamed in chunk of 3MB. This is important to keep this below 4MB as GRPC has hard limit of 4MB
	// per message
	chunkSize = 3 * 1024 * 1024
)

// sendMessagesToStream is a helper function that takes as input the payload to be sent, split it into fixed-size chunks
// and send it via the stream using sendViaStreamFunc.
// grpcMsgConstructor must be implemented by the user of this function. it should produce an actual proto message object
// from the previousChunkHash and the byte array for this chunk.
func sendMessagesToStream[DataChunkProtoMessage interface{}](
	payloadNameForLogging string,
	payload []byte,
	sendViaStreamFunc func(msg interface{}) error,
	grpcMsgConstructor func(previousChunkHash string, contentChunk []byte) (*DataChunkProtoMessage, error),
) error {
	var previousChunkHash string
	hasher := sha1.New()
	chunkNumber := 0
	totalChunksNumber := int(math.Ceil(float64(len(payload)) / chunkSize))

	for blockIdx := 0; blockIdx < totalChunksNumber; blockIdx++ {
		logrus.Debugf("Sending content for %s. Block number %d/%d", payloadNameForLogging, chunkNumber, totalChunksNumber-1)

		var contentChunk []byte
		payloadStartIdx := blockIdx * chunkSize
		if blockIdx == totalChunksNumber-1 {
			// last chunk - copy everything that's left in the payload array
			contentChunk = make([]byte, len(payload)-payloadStartIdx)
			copy(contentChunk, payload[payloadStartIdx:])
		} else {
			// copy chunkSize bytes from the content array to the contentChunk
			contentChunk = make([]byte, chunkSize)
			payloadNextStartIdx := (blockIdx + 1) * chunkSize
			copy(contentChunk, payload[payloadStartIdx:payloadNextStartIdx])
		}

		message, err := grpcMsgConstructor(previousChunkHash, contentChunk)
		if err != nil {
			return stacktrace.Propagate(err, "An unexpected error occurred assembling GRPC message from data chunk"+
				" for %s", payloadNameForLogging)
		}

		err = sendViaStreamFunc(message) // TODO: we can add a retryer here
		if err != nil {
			return stacktrace.Propagate(err, "An unexpected error occurred sending '%s'", payloadNameForLogging)
		}

		hasher.Reset()
		hasher.Write(contentChunk)
		previousChunkHash = hex.EncodeToString(hasher.Sum(nil))
	}
	return nil
}

// readMessagesFromStream is a helper function that reads the content of the stream until it returns an io.EOF.
// It extracts the valuable information (i.e. the byte array and the previous chunk hash) form the generic proto message
// using the provided grpcMsgExtractor function.
// It returns a simple byte array corresponding to the assembled payload (concatenation of all chunks)
func readMessagesFromStream[DataChunkMessageType interface{}](
	payloadNameForLogging string,
	readMsgFromStream func(msg interface{}) error,
	grpcMsgExtractor func(dataChunk *DataChunkMessageType) ([]byte, string, error),
) ([]byte, error) {
	var assembledContent []byte
	var blockIdx int
	var computedPreviousBlockHash string
	hasher := sha1.New()

	dataChunk := new(DataChunkMessageType)
	errorReceivingChunk := readMsgFromStream(dataChunk)
	for errorReceivingChunk == nil {
		chunkContent, previousChunkHashFromChunk, err := grpcMsgExtractor(dataChunk)
		if err != nil {
			return nil, stacktrace.NewError("An unexpected error occurred extracting data from the streamed GRPC "+
				"message for '%s'", payloadNameForLogging)
		}
		logrus.Debugf("Receiving content for '%s'. Block number %d", payloadNameForLogging, blockIdx)

		if previousChunkHashFromChunk != computedPreviousBlockHash {
			return nil, stacktrace.NewError("An unexpected error occurred receiving data chunk for '%s'. Hash "+
				"validation did not pass: was '%s' - wanted '%s'. Maybe a block failed to be sent (networking issues) "+
				"and now entire chain is broken. Retrying the operation might fix this issue.",
				payloadNameForLogging, previousChunkHashFromChunk, computedPreviousBlockHash)
		}
		assembledContent = append(assembledContent, chunkContent...)

		hasher.Reset()
		hasher.Write(chunkContent)
		computedPreviousBlockHash = hex.EncodeToString(hasher.Sum(nil))

		dataChunk = new(DataChunkMessageType)
		errorReceivingChunk = readMsgFromStream(dataChunk) // TODO: we can add a retryer here
	}
	if errorReceivingChunk != io.EOF {
		return nil, stacktrace.Propagate(errorReceivingChunk, "An unexpected error occurred receiving '%s'",
			payloadNameForLogging)
	}
	return assembledContent, nil
}
