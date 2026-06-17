package query

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc/dto"

	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

type IGetUserByIDQuery decorator.QueryHandler[*userdtos.GetUserByIDReq, *userdtos.UserDTO]

type getUserByIDQuery struct {
	repo           user.Repository
	productAdapter productsvc.ServiceAdapter
}

func NewGetUserByIDQuery(repo user.Repository, productAdapter productsvc.ServiceAdapter) IGetUserByIDQuery {
	return &getUserByIDQuery{repo: repo, productAdapter: productAdapter}
}

func (q *getUserByIDQuery) Handle(ctx context.Context, req *userdtos.GetUserByIDReq) (*userdtos.UserDTO, error) {
	model, err := q.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, svcerr.ErrRecordNotFound
		}
		slog.Error("get user failed", slog.Uint64("user_id", uint64(req.ID)), slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	result, err := q.productAdapter.GetListOfProducts(ctx, &dto.GetListProductReq{Page: 0, Size: 100})
	if err != nil {
		slog.Error("get products from product service failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	resp := userdtos.NewUserDTO(model)

	if len(result.Products) > 0 {
		var list []*struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			PackageID string `json:"package_id"`
		}
		for _, product := range result.Products {
			list = append(list, &struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				PackageID string `json:"package_id"`
			}{
				ID:        product.ID,
				Name:      product.Name,
				PackageID: product.PackageID,
			})
		}

		resp.Products = list
	}

	return resp, nil
}
