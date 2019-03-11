package storage

import (
	"github.com/dgraph-io/badger"
	"time"
)

type badgerDb struct {
	db *badger.DB
}

func (s *badgerDb) Len() int {
	count := 0

	// meh.
	s.Keys(func(key string) bool {
		count++
		return true
	})

	return count
}

func (s *badgerDb) Set(key string, value string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (s *badgerDb) SetExpiring(key string, value string, lifetime time.Duration) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetWithTTL([]byte(key), []byte(value), lifetime)
	})
}

func (s *badgerDb) Delete(key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (s *badgerDb) Get(key string) (string, uint64, error) {
	var value []byte
	expiresAt := uint64(0)
	now := uint64(time.Now().Unix())

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return KeyNotFound
		}
		if err != nil {
			return err
		}

		expiresAt = item.ExpiresAt()
		if expiresAt != 0 && now >= expiresAt {
			return KeyExpired
		}

		if item.IsDeletedOrExpired() {
			return KeyNotFound
		}

		value, err = item.ValueCopy(nil)

		return err
	})

	return string(value), expiresAt, err
}

func (s *badgerDb) Keys(callback func (key string) bool) {
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			keepGoing := callback(string(item.Key()))
			if !keepGoing {
				break
			}
		}
		return nil
	})
}

func NewBadgerDb(logger badger.Logger, storagePath string) (Store, error) {
	opts := badger.DefaultOptions
	opts.Dir = storagePath
	opts.ValueDir = storagePath
	opts.SyncWrites = false
	opts.Logger = logger

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &badgerDb{
		db: db,
	}, nil
}
