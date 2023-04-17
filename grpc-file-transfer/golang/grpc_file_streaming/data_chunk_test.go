package grpc_file_streaming

// TestDataChunk is a simple object faking a proto Message that would be sent via the stream in real life use of this
// package.
type TestDataChunk struct {
	Chunk             []byte
	PreviousChunkHash string
}

func NewTestDataChunk(chunk []byte, previousChunkHash string) *TestDataChunk {
	return &TestDataChunk{
		Chunk:             chunk,
		PreviousChunkHash: previousChunkHash,
	}
}
