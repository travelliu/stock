package services_test

import (
	"context"
	"stock/pkg/stockd/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueListRevoke(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()
	plain, tok, err := svc.Issue(ctx, services.IssueInput{UserID: 1, Name: "cli"})
	require.NoError(t, err)
	assert.NotEmpty(t, plain)
	assert.Equal(t, "cli", tok.Name)
	assert.Empty(t, tok.PlainOnce, "Issue returns plain via the first return; DTO must omit it after the call")

	got, err := svc.ListTokens(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	require.NoError(t, svc.RevokeToken(ctx, 1, tok.ID))
	got, _ = svc.ListTokens(ctx, 1)
	assert.Len(t, got, 0)
}

func TestIssueWithExpiry(t *testing.T) {
	svc := newService(t)
	exp := time.Now().Add(24 * time.Hour)
	_, tok, err := svc.Issue(context.Background(), services.IssueInput{UserID: 1, Name: "tmp", ExpiresAt: &exp})
	require.NoError(t, err)
	require.NotNil(t, tok.ExpiresAt)
	assert.WithinDuration(t, exp, *tok.ExpiresAt, time.Second)
}
