package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupWalletSummaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	originalDB := DB
	originalLogDB := LOG_DB

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &Log{}))

	DB = db
	LOG_DB = db

	t.Cleanup(func() {
		DB = originalDB
		LOG_DB = originalLogDB
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func TestGetUserWalletSummarySubtractsRefunds(t *testing.T) {
	db := setupWalletSummaryTestDB(t)

	require.NoError(t, db.Create(&User{
		Id:        42,
		Username:  "wallet-user",
		Password:  "password",
		UsedQuota: 1000,
	}).Error)
	require.NoError(t, db.Create(&Log{UserId: 42, Type: LogTypeRefund, Quota: 250}).Error)
	require.NoError(t, db.Create(&Log{UserId: 42, Type: LogTypeRefund, Quota: 50}).Error)
	require.NoError(t, db.Create(&Log{UserId: 42, Type: LogTypeConsume, Quota: 999}).Error)

	summary, err := GetUserWalletSummary(42)

	require.NoError(t, err)
	require.Equal(t, 1000, summary.UsedQuota)
	require.Equal(t, 300, summary.RefundedQuota)
	require.Equal(t, 700, summary.ActualUsedQuota)
}

func TestGetUserWalletSummaryDoesNotReturnNegativeActualUsage(t *testing.T) {
	db := setupWalletSummaryTestDB(t)

	require.NoError(t, db.Create(&User{
		Id:        43,
		Username:  "wallet-user-2",
		Password:  "password",
		UsedQuota: 100,
	}).Error)
	require.NoError(t, db.Create(&Log{UserId: 43, Type: LogTypeRefund, Quota: 200}).Error)

	summary, err := GetUserWalletSummary(43)

	require.NoError(t, err)
	require.Equal(t, 100, summary.UsedQuota)
	require.Equal(t, 200, summary.RefundedQuota)
	require.Equal(t, 0, summary.ActualUsedQuota)
}
