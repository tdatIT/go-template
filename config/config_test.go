package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	require.Equal(t, "/config/config", getDefaultConfig())
}

func TestNewConfigReadsFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config")

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "go-template", cfg.Server.Name)
	require.Equal(t, ":8080", cfg.Server.Port)
}
