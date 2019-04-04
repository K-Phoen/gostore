package storage

import (
	"fmt"
	logging "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type storageTestSuite struct {
	suite.Suite

	syncMap Store

	badger Store
	badgerStoragePath string
}

func (suite *storageTestSuite) SetupSuite() {
	logger, _ := logging.NewNullLogger()

	suite.syncMap = NewSyncMap(logger)

	suite.badgerStoragePath = "/tmp/gostore-test-badger-store"
	badger, err := NewBadgerDb(logger, suite.badgerStoragePath)
	if err != nil {
		panic(fmt.Sprintf("Could not start badger storage engine: %s", err))
	}
	suite.badger = badger
}

func (suite *storageTestSuite) TearDownSuite() {
	os.RemoveAll(suite.badgerStoragePath)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(storageTestSuite))
}

func (suite *storageTestSuite) TestCommonStorageFeaturesWithSyncMap() {
	commonStorageFeaturesAssertions(suite.T(), suite.syncMap)
}

func (suite *storageTestSuite) TestCommonStorageFeaturesWithBadger() {
	commonStorageFeaturesAssertions(suite.T(), suite.badger)
}

func commonStorageFeaturesAssertions(t *testing.T, store Store) {
	require.Equal(t, 0, store.Len(), "An empty store should have no length")

	store.Set("known-key", "some-value")

	require.Equal(t, 1, store.Len(), "Length should be correct for a single element")

	val, _, err := store.Get("unknown-key")
	require.Equal(t, "", val, "Getting an unknown key should return an empty string")
	require.Error(t, err, "Getting an unknown key should return an error")

	val, _, err = store.Get("known-key")
	require.Equal(t, "some-value", val, "Getting a known key should return the right value")
	require.NoError(t, err, "Getting a known key should return no error")

	store.Delete("known-key")
	val, _, err = store.Get("known-key")
	require.Equal(t, 0, store.Len(), "Length should be updated after removing an element")
	require.Equal(t, "", val, "Getting a deleted key should return an empty string")
	require.Error(t, err, "Getting a deleted key should return an error")

	lifetime, _ := time.ParseDuration("1s")
	store.SetExpiring("expiring-key", "some-value", lifetime)
	require.Equal(t, 1, store.Len(), "Length should be updated after adding an element")

	val, _, err = store.Get("expiring-key")
	require.Equal(t, "some-value", val, "Getting a non-expired key should return its value")
	require.NoError(t, err, "Getting a known, non-expired key should return no error")

	// wait for the key to expire
	time.Sleep(lifetime)

	val, _, err = store.Get("expiring-key")
	require.Equal(t, 0, store.Len(), "Length should be updated after the element has expired")
	require.Equal(t, "", val, "Getting an expired key should return an empty string")
	require.Error(t, err, "Getting an expired key should return an error")

	store.Set("some-known-key", "some-value-1")
	store.Set("some-other-key", "some-value-2")
	store.Set("yet-another-key", "some-value-3")

	require.Equal(t, 3, store.Len(), "Length should be correct")

	var allKeys []string
	store.Keys(func (key string) bool {
		allKeys = append(allKeys, key)

		return true
	})

	require.Equal(t, 3, len(allKeys), "The right number of keys should have been fetched")
	require.ElementsMatch(t, allKeys, []string{"some-known-key", "yet-another-key", "some-other-key"}, "All the keys should have been fetched")

	var twoKeys []string
	store.Keys(func (key string) bool {
		twoKeys = append(twoKeys, key)

		return len(twoKeys) < 2
	})

	require.Equal(t, 2, len(twoKeys), "Just two keys should have been fetched")
}