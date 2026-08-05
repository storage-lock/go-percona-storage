package percona_storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

// PerconaStorage 基于 Percona Server 的存储引擎实现
// Percona 完全兼容 MySQL，复用 SQL92 实现
type PerconaStorage struct {
	*sql_based_storage.SqlBasedStorage
	options *PerconaStorageOptions
}

func NewPerconaStorage(ctx context.Context, options *PerconaStorageOptions) (*PerconaStorage, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}

	baseStorage, err := sql_based_storage.NewSqlBasedStorage(&sql_based_storage.SqlBasedStorageOptions{
		ConnectionManager: options.ConnectionManager,
		SqlProvider:       sql_based_storage.NewSql92Provider(),
		TableFullName:     options.TableName,
	})
	if err != nil {
		return nil, err
	}

	s := &PerconaStorage{
		SqlBasedStorage: baseStorage,
		options:         options,
	}

	if err := s.Init(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

const StorageName = "percona-storage"

func (x *PerconaStorage) GetName() string {
	return StorageName
}

func (x *PerconaStorage) Capabilities() []storage.StorageCapability {
	return []storage.StorageCapability{
		storage.CapabilityCAS,
		storage.CapabilityReliableTime,
	}
}

func (x *PerconaStorage) CreateWithVersion(ctx context.Context, lockId string, version storage.Version, lockInformation *storage.LockInformation) error {
	err := x.SqlBasedStorage.CreateWithVersion(ctx, lockId, version, lockInformation)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			return storage_lock.ErrLockAlreadyExists
		}
	}
	return err
}

type PerconaStorageOptions struct {
	ConnectionManager storage.ConnectionManager[*sql.DB]
	TableName         string
}

func NewPerconaStorageOptions() *PerconaStorageOptions {
	return &PerconaStorageOptions{TableName: storage.DefaultStorageTableName}
}

func (x *PerconaStorageOptions) Check() error {
	if x.ConnectionManager == nil {
		return ErrConnectionManagerNil
	}
	return nil
}

var ErrConnectionManagerNil = errors.New("ConnectionManager cannot be nil")