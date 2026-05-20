package query

import (
	"context"
	"log/slog"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type IListUsersQuery decorator.QueryHandler[*userdtos.ListUsersReq, *userdtos.ListUsersRes]

type listUsersQuery struct {
	repo user.Repository
}

func NewListUsersQuery(repo user.Repository) IListUsersQuery {
	return &listUsersQuery{repo: repo}
}

func (q *listUsersQuery) Handle(ctx context.Context, req *userdtos.ListUsersReq) (*userdtos.ListUsersRes, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	users, total, err := q.repo.FindAndCount(ctx, limit, offset)
	if err != nil {
		slog.Error("list users failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	return &userdtos.ListUsersRes{
		Items: userdtos.NewUserDTOs(users),
		Total: total,
	}, nil
}
