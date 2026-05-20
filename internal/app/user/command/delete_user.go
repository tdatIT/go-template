package command

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

type IDeleteUserCommand decorator.CommandHandler[*userdtos.DeleteUserReq]

type deleteUserCommand struct {
	repo user.Repository
}

func NewDeleteUserCommand(repo user.Repository) IDeleteUserCommand {
	return &deleteUserCommand{repo: repo}
}

func (c *deleteUserCommand) Handle(ctx context.Context, req *userdtos.DeleteUserReq) error {
	if req.ID == 0 {
		return svcerr.ErrInvalidIdParam
	}

	if err := c.repo.Delete(ctx, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return svcerr.ErrRecordNotFound
		}
		slog.Error("delete user failed", slog.Uint64("user_id", uint64(req.ID)), slog.String("error", err.Error()))
		return svcerr.ErrInternalServer
	}

	return nil
}
