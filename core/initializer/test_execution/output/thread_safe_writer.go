/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package output

import (
	"io"
	"sync"
)

// A class that wraps an underlying writer, such that only one thing can be writing to it at a time
type threadSafeWriter struct {
	lock *sync.Mutex

	output io.Writer
}

func newThreadSafeWriter(output io.Writer) *threadSafeWriter {
	return &threadSafeWriter{
		lock: &sync.Mutex{},
		output: output,
	}
}

func (c *threadSafeWriter) Write(bytes []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.output.Write(bytes)
}

// We implement this to make sure that a copy from the reader to this writer is still atomic
// This works because io.Copy will call down to ReadFrom, if it exists
func (c *threadSafeWriter) ReadFrom(reader io.Reader) (n int64, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return io.Copy(c.output, reader)
}