package gostore

import (
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	KeyNotFound = errors.New("key not found")
	KeyExpired = errors.New("key has expired")
)

type entry struct {
	value string

	expiration int64
}

type Store struct {
	mutex sync.RWMutex

	data map[string]entry
}

func (e entry) Expired() bool {
	if e.expiration == 0 {
		return false
	}

	return time.Now().Unix() >= e.expiration
}

func (store *Store) Len() int {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	return len(store.data)
}

func (store *Store) Set(key string, value string) {
	store.mutex.Lock()
	store.data[key] = entry{value: value, expiration: 0}
	store.mutex.Unlock()
}

func (store *Store) SetExpiring(key string, value string, lifetime time.Duration) {
	store.mutex.Lock()
	store.data[key] = entry{value: value, expiration: time.Now().Add(lifetime).Unix()}
	store.mutex.Unlock()
}

func (store *Store) Delete(key string) {
	store.mutex.Lock()
	delete(store.data, key)
	store.mutex.Unlock()
}

func (store *Store) Get(key string) (string, int64, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	item, exists := store.data[key]

	if !exists {
		return "", 0, KeyNotFound
	}


	if item.Expired() {
		delete(store.data, key)
		return "", 0, KeyExpired
	}

	return item.value, item.expiration, nil
}

func (store *Store) Keys(callback func (key string) bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for key := range store.data {
		keepGoing := callback(key)

		if !keepGoing {
			break
		}
	}
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]entry),
	}
}
