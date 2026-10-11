package percona_storage

import (
	"testing"

	"github.com/storage-lock/go-storage"
	"github.com/stretchr/testify/assert"
)

func TestNewPerconaStorageOptions(t *testing.T) {
	opts := NewPerconaStorageOptions()
	assert.NotNil(t, opts)
	assert.Equal(t, storage.DefaultStorageTableName, opts.TableName)
	assert.Nil(t, opts.ConnectionManager)
}

func TestPerconaStorageOptions_Check_NilConnectionManager(t *testing.T) {
	opts := NewPerconaStorageOptions()
	err := opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ConnectionManager")
}

func TestPerconaStorage_GetName(t *testing.T) {
	s := &PerconaStorage{}
	assert.Equal(t, StorageName, s.GetName())
}

func TestPerconaStorage_Capabilities(t *testing.T) {
	s := &PerconaStorage{}
	caps := s.Capabilities()
	assert.Contains(t, caps, storage.CapabilityCAS)
	assert.Contains(t, caps, storage.CapabilityReliableTime)
}
