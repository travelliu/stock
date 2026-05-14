package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services/user"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestCreateAndAuthenticate(t *testing.T) {
	svc := user.New(openDB(t))
	ctx := context.Background()
	u, err := svc.Create(ctx, user.CreateInput{Username: "alice", Password: "hunter2", Role: "user"})
	require.NoError(t, err)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "user", u.Role)

	got, err := svc.Authenticate(ctx, "alice", "hunter2")
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)

	_, err = svc.Authenticate(ctx, "alice", "wrong")
	assert.Error(t, err)
}

func TestChangePassword(t *testing.T) {
	svc := user.New(openDB(t))
	ctx := context.Background()
	u, _ := svc.Create(ctx, user.CreateInput{Username: "bob", Password: "old", Role: "user"})
	require.NoError(t, svc.ChangePassword(ctx, u.ID, "old", "new"))
	assert.Error(t, svc.ChangePassword(ctx, u.ID, "wrong", "other"))
	_, err := svc.Authenticate(ctx, "bob", "new")
	require.NoError(t, err)
}

func TestDisableBlocksAuth(t *testing.T) {
	svc := user.New(openDB(t))
	ctx := context.Background()
	u, _ := svc.Create(ctx, user.CreateInput{Username: "c", Password: "p", Role: "user"})
	require.NoError(t, svc.SetDisabled(ctx, u.ID, true))
	_, err := svc.Authenticate(ctx, "c", "p")
	assert.ErrorIs(t, err, user.ErrDisabled)
}

func TestUniqueUsername(t *testing.T) {
	svc := user.New(openDB(t))
	ctx := context.Background()
	_, err := svc.Create(ctx, user.CreateInput{Username: "dup", Password: "x", Role: "user"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, user.CreateInput{Username: "dup", Password: "y", Role: "user"})
	assert.Error(t, err)
}
