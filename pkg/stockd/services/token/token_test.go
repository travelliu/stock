package token_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services/token"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestIssueListRevoke(t *testing.T) {
	svc := token.New(openDB(t))
	ctx := context.Background()
	plain, tok, err := svc.Issue(ctx, token.IssueInput{UserID: 1, Name: "cli"})
	require.NoError(t, err)
	assert.NotEmpty(t, plain)
	assert.Equal(t, "cli", tok.Name)
	assert.Empty(t, tok.PlainOnce, "Issue returns plain via the first return; DTO must omit it after the call")

	got, err := svc.List(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	require.NoError(t, svc.Revoke(ctx, 1, tok.ID))
	got, _ = svc.List(ctx, 1)
	assert.Len(t, got, 0)
}

func TestIssueWithExpiry(t *testing.T) {
	svc := token.New(openDB(t))
	exp := time.Now().Add(24 * time.Hour)
	_, tok, err := svc.Issue(context.Background(), token.IssueInput{UserID: 1, Name: "tmp", ExpiresAt: &exp})
	require.NoError(t, err)
	require.NotNil(t, tok.ExpiresAt)
	assert.WithinDuration(t, exp, *tok.ExpiresAt, time.Second)
}
