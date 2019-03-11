package storage

import (
	"github.com/pkg/errors"
	"time"
)

var (
	KeyNotFound = errors.New("key not found")
	KeyExpired = errors.New("key has expired")
)

type Store interface {
	Get(key string) (string, uint64, error)

	Set(key string, value string) error
	SetExpiring(key string, value string, lifetime time.Duration) error

	Delete(key string) error

	Len() int
	Keys(callback func (key string) bool)
}