package grpc_file_streaming

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

const (
	testFileName = "test-file"
)

func TestReadMessagesFromStream_serverStream_success(t *testing.T) {
	serverStream := NewMockServerStream(
		NewTestDataChunk([]byte("hello "), ""),
		NewTestDataChunk([]byte("world "), "c4d871ad13ad00fde9a7bb7ff7ed2543aec54241"),
		NewTestDataChunk([]byte("! =)"), "7b32ed4d82732b720681291852f38d511fef276e"),
	)

	chunkExtractor := func(dataChunk *TestDataChunk) ([]byte, string, error) {
		return dataChunk.Chunk, dataChunk.PreviousChunkHash, nil
	}

	_, err := readMessagesFromStream[TestDataChunk](testFileName, serverStream.RecvMsg, chunkExtractor)
	require.NoError(t, err)
	assembledContent, err := serverStream.GetAssembledContent()
	require.NoError(t, err)
	require.Equal(t, "hello world ! =)", string(assembledContent))
}

func TestReadMessagesFromStream_serverStream_failureReceivingChunk(t *testing.T) {
	serverStream := NewMockServerStream(
		NewTestDataChunk([]byte("hello "), ""),
		nil, // nil chunk will error in MockServerStream when receiving this chunk
		NewTestDataChunk([]byte("! =)"), "7b32ed4d82732b720681291852f38d511fef276e"),
	)

	chunkExtractor := func(dataChunk *TestDataChunk) ([]byte, string, error) {
		return dataChunk.Chunk, dataChunk.PreviousChunkHash, nil
	}

	_, err := readMessagesFromStream[TestDataChunk](testFileName, serverStream.RecvMsg, chunkExtractor)
	require.NotNil(t, err)
	expectedErr := fmt.Sprintf("An unexpected error occurred receiving '%s'", testFileName)
	require.Contains(t, err.Error(), expectedErr)
}

func TestReadMessagesFromStream_serverStream_failureWrongHash(t *testing.T) {
	serverStream := NewMockServerStream(
		NewTestDataChunk([]byte("hello "), ""),
		NewTestDataChunk([]byte("world "), "this_is_not_the_correct_hash_for_block_0"),
		NewTestDataChunk([]byte("! =)"), "7b32ed4d82732b720681291852f38d511fef276e"),
	)

	chunkExtractor := func(dataChunk *TestDataChunk) ([]byte, string, error) {
		return dataChunk.Chunk, dataChunk.PreviousChunkHash, nil
	}

	_, err := readMessagesFromStream[TestDataChunk](testFileName, serverStream.RecvMsg, chunkExtractor)
	require.NotNil(t, err)
	expectedErr := fmt.Sprintf("An unexpected error occurred receiving data chunk for '%s'. Hash validation did not pass", testFileName)
	require.Contains(t, err.Error(), expectedErr)
}

func TestReadMessagesFromStream_clientStream_success(t *testing.T) {
	clientStream := NewMockClientStream(
		NewTestDataChunk([]byte("hello "), ""),
		NewTestDataChunk([]byte("world "), "c4d871ad13ad00fde9a7bb7ff7ed2543aec54241"),
		NewTestDataChunk([]byte("! =)"), "7b32ed4d82732b720681291852f38d511fef276e"),
	)

	chunkExtractor := func(dataChunk *TestDataChunk) ([]byte, string, error) {
		return dataChunk.Chunk, dataChunk.PreviousChunkHash, nil
	}

	_, err := readMessagesFromStream[TestDataChunk](testFileName, clientStream.RecvMsg, chunkExtractor)
	require.NoError(t, err)
	require.NoError(t, clientStream.CloseSend())
	assembledContent, err := clientStream.GetAssembledContent()
	require.NoError(t, err)
	require.Equal(t, "hello world ! =)", string(assembledContent))
}

func TestSendBytesStream_successOneChunk(t *testing.T) {
	clientStream := NewMockClientStream()

	fullContent, err := generateRandomByteArray(chunkSize / 2) // half a chunk
	require.Nil(t, err)
	dataChunkConstructor := func(previousChunkHash string, contentChunk []byte) (*TestDataChunk, error) {
		return NewTestDataChunk(contentChunk, previousChunkHash), nil
	}

	err = sendMessagesToStream[TestDataChunk](testFileName, fullContent, clientStream.SendMsg, dataChunkConstructor)
	require.NoError(t, err)
	require.NoError(t, clientStream.CloseSend())
	assembledContent, err := clientStream.GetAssembledContent()
	require.NoError(t, err)
	// When this assertion fails, to help debugging, change the value of chunkSize to something reasonable, like 5.
	require.Equal(t, fullContent, assembledContent)
}

func TestSendBytesStream_successMultipleChunks(t *testing.T) {
	clientStream := NewMockClientStream()

	fullContent, err := generateRandomByteArray(chunkSize*3 + chunkSize/2) // 3 full chunks plus half a chunk
	require.Nil(t, err)
	dataChunkConstructor := func(previousChunkHash string, contentChunk []byte) (*TestDataChunk, error) {
		return NewTestDataChunk(contentChunk, previousChunkHash), nil
	}

	err = sendMessagesToStream[TestDataChunk](testFileName, fullContent, clientStream.SendMsg, dataChunkConstructor)
	require.NoError(t, err)
	require.NoError(t, clientStream.CloseSend())
	assembledContent, err := clientStream.GetAssembledContent()
	require.NoError(t, err)
	// When this assertion fails, to help debugging, change the value of chunkSize to something reasonable, like 5.
	require.Equal(t, fullContent, assembledContent)
}

//////////////////////////// HELPER FUNCTIONS BELOW \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

func generateRandomByteArray(size int) ([]byte, error) {
	randomByteStream := make([]byte, size)
	_, err := rand.Read(randomByteStream)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to generate random stream")
	}
	return randomByteStream, nil
}
