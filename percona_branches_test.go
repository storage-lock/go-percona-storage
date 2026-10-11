package percona_storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
)

type takeErrorManager struct{}

func (takeErrorManager) Name() string { return "take-error" }
func (takeErrorManager) Take(context.Context) (*sql.DB, error) {
	return nil, errors.New("take failed")
}
func (takeErrorManager) Return(context.Context, *sql.DB) error { return nil }
func (takeErrorManager) Shutdown(context.Context) error        { return nil }

func TestNewPerconaStorage_NilOptions(t *testing.T) {
	assert.Panics(t, func() { _, _ = NewPerconaStorage(context.Background(), nil) })
}

func TestNewPerconaStorage_CheckError(t *testing.T) {
	_, err := NewPerconaStorage(context.Background(), NewPerconaStorageOptions())
	assert.ErrorIs(t, err, ErrConnectionManagerNil)
}

func TestNewPerconaStorage_BaseError(t *testing.T) {
	opts := NewPerconaStorageOptions()
	opts.ConnectionManager = storage.NewFixedSqlDBConnectionManager(&sql.DB{})
	opts.TableName = ""
	_, err := NewPerconaStorage(context.Background(), opts)
	assert.Error(t, err)
}

func TestNewPerconaStorage_InitError(t *testing.T) {
	_, err := NewPerconaStorage(context.Background(), &PerconaStorageOptions{ConnectionManager: takeErrorManager{}, TableName: "locks"})
	assert.EqualError(t, err, "take failed")
}

func TestNewPerconaStorage_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	s, err := NewPerconaStorage(context.Background(), &PerconaStorageOptions{
		ConnectionManager: storage.NewFixedSqlDBConnectionManager(db),
		TableName:         "locks",
	})
	assert.NoError(t, err)
	assert.Equal(t, StorageName, s.GetName())
}

func TestPerconaStorage_CreateWithVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	s, err := NewPerconaStorage(context.Background(), &PerconaStorageOptions{
		ConnectionManager: storage.NewFixedSqlDBConnectionManager(db),
		TableName:         "locks",
	})
	assert.NoError(t, err)
	info := &storage.LockInformation{OwnerId: "owner"}

	mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("Error 1062: Duplicate entry"))
	assert.ErrorIs(t, s.CreateWithVersion(context.Background(), "lock", 1, info), storage_lock.ErrLockAlreadyExists)
	mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("other"))
	assert.EqualError(t, s.CreateWithVersion(context.Background(), "lock", 1, info), "other")
	mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
	assert.NoError(t, s.CreateWithVersion(context.Background(), "lock", 1, info))
}

func TestPerconaStorage_Watch_EmptyLockID(t *testing.T) {
	_, err := (&PerconaStorage{}).Watch(context.Background(), "")
	assert.Error(t, err)
}

func TestPerconaStorageOptions_Check_Valid(t *testing.T) {
	opts := NewPerconaStorageOptions()
	opts.ConnectionManager = storage.NewFixedSqlDBConnectionManager(&sql.DB{})
	assert.NoError(t, opts.Check())
}
