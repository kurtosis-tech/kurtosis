package streaming

import (
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Streamer[T any] interface {
	Close()
	AttachStream(stream grpc.ClientStream)
	WaitAndConsumeAll() []T
	Consume(consumer func(*T))
}

type StreamerPool[T any] struct {
	pool map[string]Streamer[T]
}

type StreamerUUID string

func NewStreamerPool[T any]() StreamerPool[T] {
	return StreamerPool[T]{
		pool: make(map[string]Streamer[T]),
	}
}

func (pool StreamerPool[T]) Add(streamer Streamer[T]) StreamerUUID {
	id := uuid.New()
	id_str := id.String()
	pool.pool[id_str] = streamer
	println("Size: %d", len(pool.pool))
	return StreamerUUID(id_str)
}

func (pool StreamerPool[T]) Consume(uuid StreamerUUID, consumer func(*T)) (bool, error) {
	println("Size: %d", len(pool.pool))
	streamer, found := pool.pool[string(uuid)]
	if !found {
		return false, nil
	} else {
		delete(pool.pool, string(uuid))
	}

	streamer.Consume(consumer)

	return true, nil
}
