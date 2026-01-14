package restapi

import (
	"sync"
	"time"
)

var (
	_engineMu        sync.RWMutex
	_engineRunning   bool
	_engineStartTime time.Time
)

func SetEngineRunning(running bool) {
	_engineMu.Lock()
	defer _engineMu.Unlock()
	_engineRunning = running
	if running {
		_engineStartTime = time.Now()
	} else {
		_engineStartTime = time.Time{}
	}
}

func IsEngineRunning() bool {
	_engineMu.RLock()
	defer _engineMu.RUnlock()
	return _engineRunning
}

func GetEngineStartTime() time.Time {
	_engineMu.RLock()
	defer _engineMu.RUnlock()
	return _engineStartTime
}
