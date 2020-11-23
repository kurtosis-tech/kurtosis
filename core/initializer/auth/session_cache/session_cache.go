/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package session_cache

// The struct to store on disk
type Session struct {
	// The actual token (which we can't trust - the user may have been able to modify this)
	Token string

	// If needed, we can store extra stuff here
}


type SessionCache interface {
	SaveSession(session Session) error
	LoadSession() (tokenResponse *Session, err error)
}

