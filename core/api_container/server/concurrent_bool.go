/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import "sync"

type ConcurrentBool struct {
	mutex *sync.Mutex
	state bool
}

func NewConcurrentBool(state bool) *ConcurrentBool {
	return &ConcurrentBool{mutex: &sync.Mutex{}, state: state}
}

func (obj *ConcurrentBool) Get() bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	return obj.state
}

func (obj *ConcurrentBool) Set(newState bool) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	obj.state = newState
}


