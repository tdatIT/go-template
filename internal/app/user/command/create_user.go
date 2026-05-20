package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

type ICreateUserCommand decorator.CommandReturnHandler[*userdtos.CreateUserReq, *userdtos.UserDTO]

type createUserCommand struct {
	repo user.Repository
}

func NewCreateUserCommand(repo user.Repository) ICreateUserCommand {
	return &createUserCommand{repo: repo}
}

func (c *createUserCommand) Handle(ctx context.Context, req *userdtos.CreateUserReq) (*userdtos.UserDTO, error) {
	model := &models.User{
		Name:  strings.TrimSpace(req.Name),
		Email: strings.TrimSpace(req.Email),
	}

	if model.Name == "" || model.Email == "" {
		return nil, svcerr.ErrInvalidParameters
	}

	if err := c.repo.Create(ctx, model); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, svcerr.ErrAlreadyExists
		}
		slog.Error("create user failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}

	return userdtos.NewUserDTO(model), nil
}
