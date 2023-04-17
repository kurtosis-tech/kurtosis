package grpc_file_streaming

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
