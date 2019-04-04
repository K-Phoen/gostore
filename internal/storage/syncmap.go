package storage

import (
	log "github.com/sirupsen/logrus"
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

	logger *log.Logger

	evictionInterval time.Duration
	// in percent
	evictionBatchSize int
}

func (e entry) Expired() bool {
	if e.expiration == 0 {
		return false
	}

	// means that sub-second lifetimes will not work as expected
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

func (m *syncMap) startEvictionRoutine() {
	ticker := time.NewTicker(m.evictionInterval)

	go func() {
		for {
			select {
			case <- ticker.C:
				m.evictExpired()
			}
		}
	}()
}

func (m *syncMap) evictExpired() {
	m.logger.Debugf("Starting eviction routine")

	evictedKeys := 0

	m.mutex.Lock()

	batchSize := int(float64(len(m.data)) * float64(m.evictionBatchSize) / 100.0)

	for key, item := range m.data {
		if item.Expired() {
			delete(m.data, key)
			evictedKeys++
		}

		if evictedKeys >= batchSize {
			break
		}
	}

	m.mutex.Unlock()

	m.logger.Debugf("Evicted %d keys (maximum batch size: %d)", evictedKeys, batchSize)
}

func NewSyncMap(logger *log.Logger) Store {
	store := &syncMap{
		data: make(map[string]entry),

		logger: logger,

		evictionInterval: 10 * time.Second,
		evictionBatchSize: 20, // percent
	}

	store.startEvictionRoutine()

	return store
}
