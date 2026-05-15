package db_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/config"
	"stock/pkg/stockd/db"
)

func TestOpen_SQLiteInMemory(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "sqlite", DSN: "file::memory:?cache=shared"}}
	gdb, err := db.Open(cfg, logrus.New())
	require.NoError(t, err)
	assert.NotNil(t, gdb)
	// Migration ran: `users` table exists.
	var n int
	require.NoError(t, gdb.Raw("SELECT count(*) FROM users").Scan(&n).Error)
	assert.Equal(t, 0, n)
}

func TestOpen_RejectsUnknownDriver(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "mssql", DSN: "x"}}
	_, err := db.Open(cfg, logrus.New())
	require.Error(t, err)
}
