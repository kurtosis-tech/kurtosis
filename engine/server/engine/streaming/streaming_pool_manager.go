package streaming

import (
	"time"

	"github.com/google/uuid"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type Streamer[T any] interface {
	Close()
	WaitAndConsumeAll() []T
	Consume(consumer func(T) error) error
	MarkForConsumption()
	IsMarkedForConsumption() bool
}

type StreamerPool[T any] struct {
	pool *expirable.LRU[StreamerUUID, *asyncStarlarkLogs]
}

type StreamerUUID string

func NewStreamerPool[T any](pool_size uint, expires_after time.Duration) StreamerPool[T] {

	pool := expirable.NewLRU(
		int(pool_size),
		func(uuid StreamerUUID, streamer *asyncStarlarkLogs) {
			logrus.Infof("Removing async log uuid %s from pool.", uuid)
			if streamer.IsMarkedForConsumption() {
				// Skipping pool exit eviction because stream is still begin consumed. Context should
				// be cancel after consumption is done.
				logrus.Debugf("Async log uuid %s is marked for consumption, skipping pool exit eviction", uuid)
				return
			}
			streamer.Close()
		},
		expires_after,
	)

	return StreamerPool[T]{
		pool: pool,
	}
}

func (streamerPool StreamerPool[T]) Contains(uuid StreamerUUID) bool {
	return streamerPool.pool.Contains(uuid)
}

func (streamerPool StreamerPool[T]) Add(streamer *asyncStarlarkLogs) StreamerUUID {
	id := uuid.New()
	id_str := StreamerUUID(id.String())
	streamerPool.pool.Add(id_str, streamer)
	return id_str
}

func (streamerPool StreamerPool[T]) Consume(uuid StreamerUUID, consumer func(*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) error) (bool, error) {
	streamer, found := streamerPool.pool.Get(uuid)
	if !found {
		return false, nil
	}

	// Mark for consumption so it doesn't get evicted (closed) when removed from
	// from the LRU cache. It'll be closed after the consumption is done (see below)
	streamer.MarkForConsumption()
	removed := streamerPool.pool.Remove(uuid)

	if removed {
		defer streamer.Close()
		if err := streamer.Consume(consumer); err != nil {
			return true, err
		}
	}

	return true, nil
}

func (streamerPool StreamerPool[T]) Clean() {
	streamerPool.pool.Purge()
}
