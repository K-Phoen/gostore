package storage

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBasicMapFeatures(t *testing.T) {
	store := NewSyncMap()

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
}