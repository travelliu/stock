package bootstrap_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
	"stock/pkg/stockd/bootstrap"
	"stock/pkg/stockd/db"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestEnsureAdmin_SeedsWhenEmpty(t *testing.T) {
	gdb := openDB(t)
	logger := logrus.New()
	plain, err := bootstrap.EnsureAdmin(gdb, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, plain, "seeded password should be returned")

	var n int64
	require.NoError(t, gdb.Model(&models.User{}).Count(&n).Error)
	assert.Equal(t, int64(1), n)

	var u models.User
	require.NoError(t, gdb.First(&u, "username = ?", "admin").Error)
	assert.Equal(t, "admin", u.Role)
}

func TestEnsureAdmin_NoopWhenUsersExist(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.User{Username: "u", PasswordHash: "h", Role: "user"}).Error)
	plain, err := bootstrap.EnsureAdmin(gdb, logrus.New())
	require.NoError(t, err)
	assert.Empty(t, plain, "should not seed when users already exist")
}
