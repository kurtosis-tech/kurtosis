package grpc_file_streaming

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc/metadata"
	"io"
)

// MockServerStream is a mock of GRPC ServerStream interface to facilitate testing the streaming capabilities this
// package implements
type MockServerStream struct {
	currentReadIdx int
	chunks         []*TestDataChunk
}

func NewEmptyMockServerStream() *MockServerStream {
	return &MockServerStream{
		currentReadIdx: 0,
		chunks:         []*TestDataChunk{},
	}
}

func NewMockServerStream(chunks ...*TestDataChunk) *MockServerStream {
	return &MockServerStream{
		currentReadIdx: 0,
		chunks:         chunks,
	}
}

func (stream *MockServerStream) SetHeader(metadata metadata.MD) error {
	panic("not implemented")
}

func (stream *MockServerStream) SendHeader(metadata metadata.MD) error {
	panic("not implemented")
}

func (stream *MockServerStream) SetTrailer(metadata metadata.MD) {
	panic("not implemented")
}

func (stream *MockServerStream) Context() context.Context {
	panic("not implemented")
}

func (stream *MockServerStream) SendMsg(msg any) error {
	castMsg, ok := msg.(*TestDataChunk)
	if !ok {
		return stacktrace.NewError("Expecting TestDataChunk for MockServerStream, got '%T'", msg)
	}
	stream.chunks = append(stream.chunks, castMsg)
	return nil
}

func (stream *MockServerStream) RecvMsg(msg any) error {
	if stream.currentReadIdx >= len(stream.chunks) {
		return io.EOF
	}
	chunk := stream.chunks[stream.currentReadIdx]
	if chunk == nil {
		return stacktrace.NewError("Error receiving chunk #%d - the chunk is nil", stream.currentReadIdx)
	}
	castMsg, ok := msg.(*TestDataChunk)
	if !ok {
		return stacktrace.NewError("Expecting a TestDataChunk for MockServerStream, got '%T'", msg)
	}
	castMsg.Chunk = chunk.Chunk
	castMsg.PreviousChunkHash = chunk.PreviousChunkHash
	stream.currentReadIdx += 1
	return nil
}

func (stream *MockServerStream) GetAssembledContent() ([]byte, error) {
	var assembledContent []byte
	for _, chunk := range stream.chunks {
		assembledContent = append(assembledContent, chunk.Chunk...)
	}
	return assembledContent, nil
}
