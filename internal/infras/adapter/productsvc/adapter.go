package productsvc

import (
	"context"

	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc/dto"
)

type ServiceAdapter interface {
	GetListOfProducts(ctx context.Context, req *dto.GetListProductReq) (*dto.GetListProductResp, error)
}
