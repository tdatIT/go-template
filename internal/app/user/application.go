package user

import (
	"github.com/tdatIT/go-template/internal/app/user/command"
	"github.com/tdatIT/go-template/internal/app/user/query"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
)

type commands struct {
	CreateUser command.ICreateUserCommand
	UpdateUser command.IUpdateUserCommand
	DeleteUser command.IDeleteUserCommand
}

type queries struct {
	GetUser   query.IGetUserByIDQuery
	ListUsers query.IListUsersQuery
}

type Application struct {
	Command commands
	Query   queries
}

func NewUserApplication(userRepo user.Repository) *Application {
	return &Application{
		Command: commands{
			CreateUser: command.NewCreateUserCommand(userRepo),
			UpdateUser: command.NewUpdateUserCommand(userRepo),
			DeleteUser: command.NewDeleteUserCommand(userRepo),
		},
		Query: queries{
			GetUser:   query.NewGetUserByIDQuery(userRepo),
			ListUsers: query.NewListUsersQuery(userRepo),
		},
	}
}
