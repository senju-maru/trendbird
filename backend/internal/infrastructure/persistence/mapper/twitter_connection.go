package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// TwitterConnectionToEntity converts a GORM TwitterConnection model to a domain entity.
func TwitterConnectionToEntity(m *model.TwitterConnection) *entity.TwitterConnection {
	return &entity.TwitterConnection{
		ID:             m.ID,
		UserID:         m.UserID,
		AccessToken:    m.AccessToken,
		RefreshToken:   m.RefreshToken,
		TokenExpiresAt: m.TokenExpiresAt,
		Status:         entity.TwitterConnectionStatus(m.Status),
		ConnectedAt:    m.ConnectedAt,
		LastTestedAt:   m.LastTestedAt,
		ErrorMessage:   m.ErrorMessage,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// TwitterConnectionToModel converts a domain TwitterConnection entity to a GORM model.
func TwitterConnectionToModel(e *entity.TwitterConnection) *model.TwitterConnection {
	return &model.TwitterConnection{
		ID:             e.ID,
		UserID:         e.UserID,
		AccessToken:    e.AccessToken,
		RefreshToken:   e.RefreshToken,
		TokenExpiresAt: e.TokenExpiresAt,
		Status:         int32(e.Status),
		ConnectedAt:    e.ConnectedAt,
		LastTestedAt:   e.LastTestedAt,
		ErrorMessage:   e.ErrorMessage,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
