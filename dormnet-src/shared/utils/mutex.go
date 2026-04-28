package utils

import "sync"

type KeyMutex interface {
	Lock(key string)
	TryLock(key string) bool
	Unlock(key string)
}

type AbsKeyMutex struct {
	mutex      sync.Mutex
	keyMutexes map[string]*sync.Mutex
}

func NewAirKVMutex() KeyMutex {
	return &AbsKeyMutex{
		keyMutexes: make(map[string]*sync.Mutex),
	}
}

func (m *AbsKeyMutex) registerMutex(key string) *sync.Mutex {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.keyMutexes[key] == nil {
		m.keyMutexes[key] = &sync.Mutex{}
	}
	mutex := m.keyMutexes[key]
	return mutex
}
func (m *AbsKeyMutex) Lock(key string) {
	mutex := m.registerMutex(key)
	mutex.Lock()
}
func (m *AbsKeyMutex) TryLock(key string) bool {
	mutex := m.registerMutex(key)
	return mutex.TryLock()
}

func (m *AbsKeyMutex) Unlock(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.keyMutexes[key] == nil {
		return
	}
	mutex := m.keyMutexes[key]
	mutex.Unlock()
	if mutex.TryLock() {
		mutex.Unlock()
		delete(m.keyMutexes, key)
	}
}
