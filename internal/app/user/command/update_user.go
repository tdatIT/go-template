package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

type IUpdateUserCommand decorator.CommandReturnHandler[*userdtos.UpdateUserReq, *userdtos.UserDTO]

type updateUserCommand struct {
	repo user.Repository
}

func NewUpdateUserCommand(repo user.Repository) IUpdateUserCommand {
	return &updateUserCommand{repo: repo}
}

func (c *updateUserCommand) Handle(ctx context.Context, req *userdtos.UpdateUserReq) (*userdtos.UserDTO, error) {
	if req.Name == nil && req.Email == nil {
		return nil, svcerr.ErrInvalidParameters
	}

	model, err := c.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, svcerr.ErrRecordNotFound
		}
		slog.Error("get user for update failed", slog.Uint64("user_id", uint64(req.ID)), slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	if req.Name != nil {
		model.Name = strings.TrimSpace(*req.Name)
	}

	if req.Email != nil {
		model.Email = strings.TrimSpace(*req.Email)
	}

	if err := c.repo.Update(ctx, model); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, svcerr.ErrAlreadyExists
		}
		slog.Error("update user failed", slog.Uint64("user_id", uint64(req.ID)), slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	return userdtos.NewUserDTO(model), nil
}
