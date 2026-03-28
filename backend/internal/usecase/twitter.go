package usecase

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
)

// TwitterUsecase handles Twitter connection management operations.
type TwitterUsecase struct {
	tweetConnRepo repository.TwitterConnectionRepository
	twitterGW     gateway.TwitterGateway
}

// NewTwitterUsecase creates a new TwitterUsecase.
func NewTwitterUsecase(
	tweetConnRepo repository.TwitterConnectionRepository,
	twitterGW gateway.TwitterGateway,
) *TwitterUsecase {
	return &TwitterUsecase{
		tweetConnRepo: tweetConnRepo,
		twitterGW:     twitterGW,
	}
}

// GetConnectionInfo returns the Twitter connection for the user.
func (u *TwitterUsecase) GetConnectionInfo(ctx context.Context, userID string) (*entity.TwitterConnection, error) {
	return u.tweetConnRepo.FindByUserID(ctx, userID)
}

// TestConnection verifies the user's Twitter credentials, refreshing the token if expired.
func (u *TwitterUsecase) TestConnection(ctx context.Context, userID string) error {
	conn, err := u.tweetConnRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	accessToken := conn.AccessToken

	// Refresh token if expired
	if conn.TokenExpiresAt.Before(time.Now()) {
		tokenResp, err := u.twitterGW.RefreshToken(ctx, conn.RefreshToken)
		if err != nil {
			errMsg := "failed to refresh token: " + err.Error()
			_ = u.tweetConnRepo.UpdateStatus(ctx, userID, entity.TwitterError, &errMsg)
			return apperror.Wrap(apperror.CodeInternal, "failed to refresh X token", err)
		}

		now := time.Now()
		updated := &entity.TwitterConnection{
			UserID:         userID,
			AccessToken:    tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
			TokenExpiresAt: now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			Status:         conn.Status,
			ConnectedAt:    conn.ConnectedAt,
		}
		if err := u.tweetConnRepo.Upsert(ctx, updated); err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to update refreshed token", err)
		}
		accessToken = tokenResp.AccessToken
	}

	// Verify credentials with X API
	if err := u.twitterGW.VerifyCredentials(ctx, accessToken); err != nil {
		errMsg := "credential verification failed: " + err.Error()
		_ = u.tweetConnRepo.UpdateStatus(ctx, userID, entity.TwitterError, &errMsg)
		return apperror.Wrap(apperror.CodeInternal, "X credential verification failed", err)
	}

	// Success: update status and last tested timestamp
	_ = u.tweetConnRepo.UpdateStatus(ctx, userID, entity.TwitterConnected, nil)
	_ = u.tweetConnRepo.UpdateLastTestedAt(ctx, userID)

	return nil
}

// DisconnectTwitter removes the user's Twitter connection.
func (u *TwitterUsecase) DisconnectTwitter(ctx context.Context, userID string) error {
	return u.tweetConnRepo.DeleteByUserID(ctx, userID)
}
