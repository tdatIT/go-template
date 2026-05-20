package svcerr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorMessage(t *testing.T) {
	err := &Error{Message: "msg"}
	require.Equal(t, "msg", err.Error())

	ret := err.WithError(errors.New("inner"))
	require.Same(t, err, ret)
	require.Equal(t, "inner", err.Error())
}
