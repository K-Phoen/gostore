package main

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

type Store struct {
	sync.RWMutex

	data map[string]string
}

func (store *Store) Set(key string, value string) {
	store.Lock()
	store.data[key] = value
	store.Unlock()
}

func (store *Store) Delete(key string) {
	store.Lock()
	delete(store.data, key)
	store.Unlock()
}

func (store *Store) Get(key string) (string, error) {
	store.RLock()
	defer store.RUnlock()

	if val, exists := store.data[key]; exists {
		return val, nil
	}

	return "", errors.New(fmt.Sprintf("No data for key %q", key))
}

func NewStore() Store {
	return Store{
		data: make(map[string]string),
	}
}