package user

import (
	"context"

	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/domain/models"
	"github.com/tdatIT/go-template/pkgs/db/orm"
)

type userReposImpl struct {
	db orm.ORM
}

func NewUserRepository(database orm.ORM) Repository {
	return &userReposImpl{db: database}
}

func (r *userReposImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.GormDB().WithContext(ctx).
		First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userReposImpl) FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var count int64

	err := r.db.GormDB().WithContext(ctx).
		Model(&models.User{}).
		Count(&count).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

func (r *userReposImpl) Create(ctx context.Context, user *models.User) error {
	err := r.db.GormDB().WithContext(ctx).
		Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userReposImpl) Update(ctx context.Context, user *models.User) error {
	err := r.db.GormDB().WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userReposImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.GormDB().WithContext(ctx).
		Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
