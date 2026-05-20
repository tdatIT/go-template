package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tdatIT/go-template/config"
	"github.com/tdatIT/go-template/pkgs/ultis/validate"
)

func TestNewHttpServer(t *testing.T) {
	cfg := &config.AppConfig{
		Server: config.Server{
			Name: "test-service",
		},
	}

	e := newHttpServer(cfg)
	require.NotNil(t, e)
	require.NotNil(t, e.HTTPErrorHandler)
	require.NotNil(t, e.Logger)
	require.Same(t, validate.GetValidator(), e.Validator)
}
