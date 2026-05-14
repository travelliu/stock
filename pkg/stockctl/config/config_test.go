package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockctl/config"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	cfg := &config.Config{ServerURL: "https://example.com", Token: "stk_test"}
	require.NoError(t, cfg.Save(p))
	got, err := config.Load(p)
	require.NoError(t, err)
	assert.Equal(t, cfg.ServerURL, got.ServerURL)
	assert.Equal(t, cfg.Token, got.Token)
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	got, err := config.Load(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Empty(t, got.ServerURL)
}
