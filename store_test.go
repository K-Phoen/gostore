package gostore

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasicFeatures(t *testing.T) {
	store := NewStore()

	store.Set("known-key", "some-value")

	val, err := store.Get("unknown-key")
	require.Equal(t, "", val, "Getting an unknown key should return an empty string")
	require.Error(t, err, "Getting an unknown key should return an error")

	val, err = store.Get("known-key")
	require.Equal(t, "some-value", val, "Getting a known key should return the right value")
	require.NoError(t, err, "Getting a known key should return no error")

	store.Delete("known-key")
	val, err = store.Get("known-key")
	require.Equal(t, "", val, "Getting a deleted key should return an empty string")
	require.Error(t, err, "Getting a deleted key should return an error")
}