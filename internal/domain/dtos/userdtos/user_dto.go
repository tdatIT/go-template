package userdtos

import (
	"time"

	"github.com/tdatIT/go-template/internal/domain/models"
)

type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
}

type UpdateUserReq struct {
	ID    uint    `json:"-"`
	Name  *string `json:"name" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email" validate:"omitempty,email,max=255"`
}

type DeleteUserReq struct {
	ID uint `json:"-"`
}

type GetUserByIDReq struct {
	ID uint `json:"-"`
}

type ListUsersReq struct {
	Limit  int
	Offset int
}

type ListUsersRes struct {
	Items []*UserDTO `json:"items"`
	Total int64      `json:"total"`
}

type UserDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserDTO(user *models.User) *UserDTO {
	if user == nil {
		return nil
	}

	return &UserDTO{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func NewUserDTOs(users []*models.User) []*UserDTO {
	if len(users) == 0 {
		return []*UserDTO{}
	}

	result := make([]*UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, NewUserDTO(user))
	}

	return result
}
