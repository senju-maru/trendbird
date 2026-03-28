package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ domainrepo.TopicRepository = (*topicRepository)(nil)

type topicRepository struct {
	db *gorm.DB
}

func NewTopicRepository(db *gorm.DB) *topicRepository {
	return &topicRepository{db: db}
}

func (r *topicRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *topicRepository) FindByID(ctx context.Context, id string) (*entity.Topic, error) {
	var m model.Topic
	if err := r.getDB(ctx).Preload("Genre").Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("topic not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find topic", err)
	}
	return mapper.TopicToEntity(&m), nil
}

// FindByIDForUser returns a topic by ID with notification_enabled from user_topics.
// Returns NotFound if the user does not have a link to the topic.
func (r *topicRepository) FindByIDForUser(ctx context.Context, id string, userID string) (*entity.Topic, error) {
	var result struct {
		model.Topic
		NotificationEnabled bool   `gorm:"column:notification_enabled"`
		GenreSlug           string `gorm:"column:genre_slug"`
	}
	err := r.getDB(ctx).
		Table("topics").
		Select("topics.*, user_topics.notification_enabled, genres.slug as genre_slug").
		Joins("JOIN user_topics ON topics.id = user_topics.topic_id").
		Joins("JOIN genres ON topics.genre_id = genres.id").
		Where("topics.id = ? AND user_topics.user_id = ?", id, userID).
		First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("topic not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find topic for user", err)
	}
	e := mapper.TopicToEntity(&result.Topic)
	e.NotificationEnabled = result.NotificationEnabled
	e.GenreSlug = result.GenreSlug
	return e, nil
}

// FindByNameAndGenre searches for a topic by (name, genre_id). Returns nil, nil if not found.
// genreID は genres テーブルの UUID。
func (r *topicRepository) FindByNameAndGenre(ctx context.Context, name string, genreID string) (*entity.Topic, error) {
	var m model.Topic
	if err := r.getDB(ctx).Preload("Genre").Where("name = ? AND genre_id = ?", name, genreID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find topic by name and genre", err)
	}
	return mapper.TopicToEntity(&m), nil
}

func (r *topicRepository) ListAll(ctx context.Context) ([]*entity.Topic, error) {
	var models []model.Topic
	if err := r.getDB(ctx).Preload("Genre").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list all topics", err)
	}
	entities := make([]*entity.Topic, len(models))
	for i := range models {
		entities[i] = mapper.TopicToEntity(&models[i])
	}
	return entities, nil
}

func (r *topicRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Topic, error) {
	type topicWithNotif struct {
		model.Topic
		NotificationEnabled bool   `gorm:"column:notification_enabled"`
		GenreSlug           string `gorm:"column:genre_slug"`
	}
	var results []topicWithNotif
	err := r.getDB(ctx).
		Table("topics").
		Select("topics.*, user_topics.notification_enabled, genres.slug as genre_slug").
		Joins("JOIN user_topics ON topics.id = user_topics.topic_id").
		Joins("JOIN genres ON topics.genre_id = genres.id").
		Where("user_topics.user_id = ?", userID).
		Find(&results).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list topics", err)
	}
	entities := make([]*entity.Topic, len(results))
	for i := range results {
		entities[i] = mapper.TopicToEntity(&results[i].Topic)
		entities[i].NotificationEnabled = results[i].NotificationEnabled
		entities[i].GenreSlug = results[i].GenreSlug
	}
	return entities, nil
}

func (r *topicRepository) GetLatestUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error) {
	var result sql.NullTime
	err := r.getDB(ctx).
		Table("topics").
		Select("MAX(topics.updated_at)").
		Joins("JOIN user_topics ON topics.id = user_topics.topic_id").
		Where("user_topics.user_id = ?", userID).
		Scan(&result).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to get latest updated_at", err)
	}
	if !result.Valid {
		return nil, nil
	}
	return &result.Time, nil
}

func (r *topicRepository) Create(ctx context.Context, topic *entity.Topic) error {
	m := mapper.TopicToModel(topic)
	if err := r.getDB(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create topic", err)
	}
	// If conflict (DO NOTHING), m.ID remains empty — fetch the existing record
	if m.ID == "" {
		if err := r.getDB(ctx).Preload("Genre").Where("name = ? AND genre_id = ?", m.Name, m.GenreID).First(m).Error; err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to find conflicting topic", err)
		}
	}
	topic.ID = m.ID
	topic.CreatedAt = m.CreatedAt
	topic.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *topicRepository) Update(ctx context.Context, topic *entity.Topic) error {
	m := mapper.TopicToModel(topic)
	if err := r.getDB(ctx).Save(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update topic", err)
	}
	topic.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *topicRepository) Delete(ctx context.Context, id string) error {
	result := r.getDB(ctx).Where("id = ?", id).Delete(&model.Topic{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete topic", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("topic not found")
	}
	return nil
}

func (r *topicRepository) SuggestByName(ctx context.Context, query string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error) {
	type row struct {
		ID              string  `gorm:"column:id"`
		Name            string  `gorm:"column:name"`
		Keywords        string  `gorm:"column:keywords"`
		GenreSlug       string  `gorm:"column:genre_slug"`
		GenreLabel      string  `gorm:"column:genre_label"`
		SimilarityScore float64 `gorm:"column:similarity_score"`
	}

	db := r.getDB(ctx).
		Table("topics").
		Select("topics.id, topics.name, topics.keywords, genres.slug as genre_slug, genres.label as genre_label, similarity(topics.name, ?) as similarity_score", query).
		Joins("JOIN genres ON topics.genre_id = genres.id").
		Where("similarity(topics.name, ?) > 0.1", query).
		Order("similarity_score DESC").
		Limit(limit)

	if len(excludeIDs) > 0 {
		db = db.Where("topics.id NOT IN ?", excludeIDs)
	}

	var rows []row
	if err := db.Find(&rows).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to suggest topics", err)
	}

	results := make([]*entity.TopicSuggestion, len(rows))
	for i, r := range rows {
		results[i] = &entity.TopicSuggestion{
			ID:              r.ID,
			Name:            r.Name,
			Keywords:        mapper.JSONToStringSlice(r.Keywords),
			GenreSlug:       r.GenreSlug,
			GenreLabel:      r.GenreLabel,
			SimilarityScore: r.SimilarityScore,
		}
	}
	return results, nil
}

func (r *topicRepository) ListByGenreExcluding(ctx context.Context, genreSlug string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error) {
	type row struct {
		ID         string `gorm:"column:id"`
		Name       string `gorm:"column:name"`
		Keywords   string `gorm:"column:keywords"`
		GenreSlug  string `gorm:"column:genre_slug"`
		GenreLabel string `gorm:"column:genre_label"`
	}

	db := r.getDB(ctx).
		Table("topics").
		Select("topics.id, topics.name, topics.keywords, genres.slug as genre_slug, genres.label as genre_label").
		Joins("JOIN genres ON topics.genre_id = genres.id").
		Where("genres.slug = ?", genreSlug).
		Order("topics.created_at DESC").
		Limit(limit)

	if len(excludeIDs) > 0 {
		db = db.Where("topics.id NOT IN ?", excludeIDs)
	}

	var rows []row
	if err := db.Find(&rows).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list topics by genre", err)
	}

	results := make([]*entity.TopicSuggestion, len(rows))
	for i, r := range rows {
		results[i] = &entity.TopicSuggestion{
			ID:         r.ID,
			Name:       r.Name,
			Keywords:   mapper.JSONToStringSlice(r.Keywords),
			GenreSlug:  r.GenreSlug,
			GenreLabel: r.GenreLabel,
		}
	}
	return results, nil
}
