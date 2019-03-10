package gostore

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

type Store struct {
	mutex sync.RWMutex

	data map[string]string
}

func (store *Store) Len() int {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	return len(store.data)
}

func (store *Store) Set(key string, value string) {
	store.mutex.Lock()
	store.data[key] = value
	store.mutex.Unlock()
}

func (store *Store) Delete(key string) {
	store.mutex.Lock()
	delete(store.data, key)
	store.mutex.Unlock()
}

func (store *Store) Get(key string) (string, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if val, exists := store.data[key]; exists {
		return val, nil
	}

	return "", errors.New(fmt.Sprintf("No data for key %q", key))
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
		data: make(map[string]string),
	}
}
