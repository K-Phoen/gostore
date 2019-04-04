package storage

import (
	logging "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type syncmapTestSuite struct {
	suite.Suite

	store syncMap
}

func (suite *syncmapTestSuite) SetupSuite() {
	logger, _ := logging.NewNullLogger()

	store := syncMap{
		data: make(map[string]entry),

		logger: logger,

		evictionInterval: 250 * time.Millisecond, // not relevant here as we trigger the eviction process manually
		evictionBatchSize: 100, // percent
	}

	suite.store = store
}

func TestSyncMapTestSuite(t *testing.T) {
	suite.Run(t, new(syncmapTestSuite))
}

func (suite *syncmapTestSuite) TestEvictionRoutine() {
	require := suite.Require()

	suite.store.Set("known-key", "some-value")

	lifetime, _ := time.ParseDuration("1s")
	suite.store.SetExpiring("expiring-key", "some-value", lifetime)

	require.Equal(2, suite.store.Len(), "Length should be correct")

	// wait for the key to expire and the eviction routine to run
	time.Sleep(lifetime * 2)

	suite.store.evictExpired()

	require.Equal(1, suite.store.Len(), "Length should be correct and the expired key should have been deleted by the eviction routine")
}