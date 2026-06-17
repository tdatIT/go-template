package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appuser "github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc"
	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc/dto"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
)

type fakeProductAdapter struct{}

func (f *fakeProductAdapter) GetListOfProducts(_ context.Context, _ *dto.GetListProductReq) (*dto.GetListProductResp, error) {
	return &dto.GetListProductResp{}, nil
}

var _ productsvc.ServiceAdapter = (*fakeProductAdapter)(nil)

func TestNewUserApplication_WiresAllHandlers(t *testing.T) {
	repo := repomock.NewMockRepository(t)

	app := appuser.NewUserApplication(repo, &fakeProductAdapter{})

	require.NotNil(t, app)
	assert.NotNil(t, app.Command.CreateUser)
	assert.NotNil(t, app.Command.UpdateUser)
	assert.NotNil(t, app.Command.DeleteUser)
	assert.NotNil(t, app.Query.GetUser)
	assert.NotNil(t, app.Query.ListUsers)
}
