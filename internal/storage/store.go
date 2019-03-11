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
	Get(key string) (string, int64, error)

	Set(key string, value string)
	SetExpiring(key string, value string, lifetime time.Duration)

	Delete(key string)

	Len() int
	Keys(callback func (key string) bool)
}