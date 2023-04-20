package grpc_file_streaming

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc/metadata"
	"io"
)

// MockClientStream is a mock of GRPC ClientStream interface to facilitate testing the streaming capabilities this
// package implements
type MockClientStream struct {
	currentReadIdx int
	chunks         []*TestDataChunk
	isClosed       bool
}

func NewEmptyMockClientStream() *MockClientStream {
	return &MockClientStream{
		currentReadIdx: 0,
		chunks:         []*TestDataChunk{},
		isClosed:       false,
	}
}

func NewMockClientStream(chunks ...*TestDataChunk) *MockClientStream {
	return &MockClientStream{
		currentReadIdx: 0,
		chunks:         chunks,
		isClosed:       false,
	}
}

func (stream *MockClientStream) Header() (metadata.MD, error) {
	panic("not implemented")
}

func (stream *MockClientStream) Trailer() metadata.MD {
	panic("not implemented")
}

func (stream *MockClientStream) CloseSend() error {
	stream.isClosed = true
	return nil
}

func (stream *MockClientStream) Context() context.Context {
	panic("not implemented")
}

func (stream *MockClientStream) SendMsg(msg any) error {
	castMsg, ok := msg.(*TestDataChunk)
	if !ok {
		return stacktrace.NewError("Expecting a TestDataChunk for MockClientStream")
	}
	stream.chunks = append(stream.chunks, castMsg)
	return nil
}

func (stream *MockClientStream) RecvMsg(msg any) error {
	if stream.currentReadIdx >= len(stream.chunks) {
		return io.EOF
	}
	chunk := stream.chunks[stream.currentReadIdx]
	if chunk == nil {
		return stacktrace.NewError("Error receiving chunk #%d", stream.currentReadIdx)
	}
	castMsg, ok := msg.(*TestDataChunk)
	if !ok {
		return stacktrace.NewError("Expecting a TestDataChunk for MockServerStream")
	}
	castMsg.Chunk = chunk.Chunk
	castMsg.PreviousChunkHash = chunk.PreviousChunkHash
	stream.currentReadIdx += 1
	return nil
}

func (stream *MockClientStream) GetAssembledContent() ([]byte, error) {
	if !stream.isClosed {
		return nil, stacktrace.NewError("Stream needs to be closed before it can assemble the full content")
	}
	var assembledContent []byte
	for _, chunk := range stream.chunks {
		assembledContent = append(assembledContent, chunk.Chunk...)
	}
	return assembledContent, nil
}
