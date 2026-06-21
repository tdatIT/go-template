package query

import (
	"context"
	"log/slog"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	pagable "github.com/tdatIT/go-template/pkgs/utilities/paging"
	"github.com/tdatIT/go-template/pkgs/utilities/svcerr"
)

type IListUsersQuery decorator.QueryHandler[*pagable.ListQuery, *pagable.ListResponse]

type listUsersQuery struct {
	repo user.Repository
}

func NewListUsersQuery(repo user.Repository) IListUsersQuery {
	return &listUsersQuery{repo: repo}
}

func (q *listUsersQuery) Handle(ctx context.Context, req *pagable.ListQuery) (*pagable.ListResponse, error) {
	users, total, err := q.repo.FindAndCount(ctx, req.GetSize(), req.GetOffset())
	if err != nil {
		slog.Error("list users failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	response := &pagable.ListResponse{
		Items:   users,
		Total:   int(total),
		Page:    req.GetPage(),
		Size:    req.GetSize(),
		HasMore: req.GetHasMore(int(total)),
	}

	return response, nil
}
