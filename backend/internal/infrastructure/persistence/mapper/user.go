package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// UserToEntity converts a GORM User model to a domain User entity.
func UserToEntity(m *model.User) *entity.User {
	return &entity.User{
		ID:                m.ID,
		TwitterID:         m.TwitterID,
		Name:              m.Name,
		Email:             m.Email,
		Image:             m.Image,
		TwitterHandle:     m.TwitterHandle,
		TutorialCompleted: m.TutorialCompleted,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// UserToModel converts a domain User entity to a GORM User model.
func UserToModel(e *entity.User) *model.User {
	return &model.User{
		ID:                e.ID,
		TwitterID:         e.TwitterID,
		Name:              e.Name,
		Email:             e.Email,
		Image:             e.Image,
		TwitterHandle:     e.TwitterHandle,
		TutorialCompleted: e.TutorialCompleted,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}
