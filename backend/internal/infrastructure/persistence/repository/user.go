package repository

import (
	"context"
	"errors"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ domainrepo.UserRepository = (*userRepository)(nil)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var m model.User
	if err := r.getDB(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find user by id", err)
	}
	return mapper.UserToEntity(&m), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.User
	if err := r.getDB(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find user by email", err)
	}
	return mapper.UserToEntity(&m), nil
}

func (r *userRepository) FindByTwitterID(ctx context.Context, twitterID string) (*entity.User, error) {
	var m model.User
	if err := r.getDB(ctx).Where("twitter_id = ?", twitterID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find user by twitter id", err)
	}
	return mapper.UserToEntity(&m), nil
}

func (r *userRepository) UpsertByTwitterID(ctx context.Context, input entity.UpsertUserInput) (*entity.User, error) {
	m := model.User{
		TwitterID:     input.TwitterID,
		Name:          input.Name,
		Email:         input.Email,
		Image:         input.Image,
		TwitterHandle: input.TwitterHandle,
	}
	db := r.getDB(ctx)
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "twitter_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "image", "twitter_handle", "updated_at",
		}),
	}).Create(&m).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to upsert user", err)
	}
	// GORM の OnConflict upsert は RETURNING "id" のみで created_at を返さないため、
	// 正確なタイムスタンプを取得するために再読み込みする。
	if err := db.First(&m, "id = ?", m.ID).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to re-read upserted user", err)
	}
	return mapper.UserToEntity(&m), nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, id string, email string) error {
	result := r.getDB(ctx).Model(&model.User{}).Where("id = ?", id).Update("email", email)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update user email", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}

func (r *userRepository) ListByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var models []model.User
	err := r.getDB(ctx).Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list users by ids", err)
	}
	entities := make([]*entity.User, len(models))
	for i := range models {
		entities[i] = mapper.UserToEntity(&models[i])
	}
	return entities, nil
}

func (r *userRepository) CompleteTutorial(ctx context.Context, id string) error {
	result := r.getDB(ctx).Model(&model.User{}).
		Where("id = ? AND tutorial_completed = false", id).
		Update("tutorial_completed", true)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to complete tutorial", result.Error)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.getDB(ctx).Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}
