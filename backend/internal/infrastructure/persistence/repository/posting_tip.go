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

var _ domainrepo.PostingTipRepository = (*postingTipRepository)(nil)

type postingTipRepository struct {
	db *gorm.DB
}

func NewPostingTipRepository(db *gorm.DB) *postingTipRepository {
	return &postingTipRepository{db: db}
}

func (r *postingTipRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *postingTipRepository) FindByTopicID(ctx context.Context, topicID string) (*entity.PostingTip, error) {
	var m model.PostingTip
	if err := r.getDB(ctx).Where("topic_id = ?", topicID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("posting tip not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find posting tip", err)
	}
	return mapper.PostingTipToEntity(&m), nil
}

func (r *postingTipRepository) Upsert(ctx context.Context, tip *entity.PostingTip) error {
	m := mapper.PostingTipToModel(tip)
	err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "topic_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"peak_days", "peak_hours_start", "peak_hours_end",
			"next_suggested_time", "updated_at",
		}),
	}).Create(m).Error
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to upsert posting tip", err)
	}
	return nil
}
