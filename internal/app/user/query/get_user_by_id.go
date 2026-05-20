package query

import (
	"context"
	"errors"
	"log/slog"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
	"gorm.io/gorm"
)

type IGetUserByIDQuery decorator.QueryHandler[*userdtos.GetUserByIDReq, *userdtos.UserDTO]

type getUserByIDQuery struct {
	repo user.Repository
}

func NewGetUserByIDQuery(repo user.Repository) IGetUserByIDQuery {
	return &getUserByIDQuery{repo: repo}
}

func (q *getUserByIDQuery) Handle(ctx context.Context, req *userdtos.GetUserByIDReq) (*userdtos.UserDTO, error) {
	if req.ID == 0 {
		return nil, svcerr.ErrInvalidIdParam
	}

	model, err := q.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, svcerr.ErrRecordNotFound
		}
		slog.Error("get user failed", slog.Uint64("user_id", uint64(req.ID)), slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	return userdtos.NewUserDTO(model), nil
}
