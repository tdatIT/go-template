package userdtos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tdatIT/go-template/internal/domain/models"
)

func TestNewUserDTO(t *testing.T) {
	require.Nil(t, NewUserDTO(nil))

	now := time.Now()
	user := &models.User{
		ID:        10,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	dto := NewUserDTO(user)
	require.NotNil(t, dto)
	require.Equal(t, user.ID, dto.ID)
	require.Equal(t, user.Name, dto.Name)
	require.Equal(t, user.Email, dto.Email)
	require.Equal(t, user.CreatedAt, dto.CreatedAt)
	require.Equal(t, user.UpdatedAt, dto.UpdatedAt)
}

func TestNewUserDTOs(t *testing.T) {
	require.Equal(t, []*UserDTO{}, NewUserDTOs(nil))

	users := []*models.User{
		{ID: 1, Name: "A", Email: "a@example.com"},
		{ID: 2, Name: "B", Email: "b@example.com"},
	}
	dtos := NewUserDTOs(users)
	require.Len(t, dtos, 2)
	require.Equal(t, uint(1), dtos[0].ID)
	require.Equal(t, uint(2), dtos[1].ID)
}
