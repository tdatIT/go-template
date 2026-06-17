package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appuser "github.com/tdatIT/go-template/internal/app/user"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
)

func TestNewUserApplication_WiresAllHandlers(t *testing.T) {
	repo := repomock.NewMockRepository(t)

	app := appuser.NewUserApplication(repo)

	require.NotNil(t, app)
	assert.NotNil(t, app.Command.CreateUser)
	assert.NotNil(t, app.Command.UpdateUser)
	assert.NotNil(t, app.Command.DeleteUser)
	assert.NotNil(t, app.Query.GetUser)
	assert.NotNil(t, app.Query.ListUsers)
}
