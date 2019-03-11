package storage

import (
	"sync"
	"time"
)

type entry struct {
	value string

	expiration uint64
}

type syncMap struct {
	mutex sync.RWMutex

	data map[string]entry
}

func (e entry) Expired() bool {
	if e.expiration == 0 {
		return false
	}

	return uint64(time.Now().Unix()) >= e.expiration
}

func (m *syncMap) Len() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.data)
}

func (m *syncMap) Set(key string, value string) error {
	m.mutex.Lock()
	m.data[key] = entry{value: value, expiration: 0}
	m.mutex.Unlock()

	return nil
}

func (m *syncMap) SetExpiring(key string, value string, lifetime time.Duration) error {
	m.mutex.Lock()
	m.data[key] = entry{value: value, expiration: uint64(time.Now().Add(lifetime).Unix())}
	m.mutex.Unlock()

	return nil
}

func (m *syncMap) Delete(key string) error {
	m.mutex.Lock()
	delete(m.data, key)
	m.mutex.Unlock()

	return nil
}

func (m *syncMap) Get(key string) (string, uint64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]

	if !exists {
		return "", 0, KeyNotFound
	}


	if item.Expired() {
		delete(m.data, key)
		return "", 0, KeyExpired
	}

	return item.value, item.expiration, nil
}

func (m *syncMap) Keys(callback func (key string) bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for key := range m.data {
		keepGoing := callback(key)

		if !keepGoing {
			break
		}
	}
}

func NewSyncMap() Store {
	return &syncMap{
		data: make(map[string]entry),
	}
}
