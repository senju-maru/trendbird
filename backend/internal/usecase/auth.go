package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/auth"
)

// XAuthResult holds the result of a successful X OAuth authentication.
type XAuthResult struct {
	User            *entity.User
	Token           string
	TutorialPending bool
}

// GetCurrentUserResult holds the result of GetCurrentUser.
type GetCurrentUserResult struct {
	User            *entity.User
	TutorialPending bool
}

// AuthUsecase handles X OAuth authentication and account management.
type AuthUsecase struct {
	userRepo        repository.UserRepository
	tweetConnRepo   repository.TwitterConnectionRepository
	notiSettingRepo repository.NotificationSettingRepository
	activityRepo    repository.ActivityRepository
	twitterGW       gateway.TwitterGateway
	jwtSvc          *auth.JWTService
	txManager       repository.TransactionManager
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(
	userRepo repository.UserRepository,
	tweetConnRepo repository.TwitterConnectionRepository,
	notiSettingRepo repository.NotificationSettingRepository,
	activityRepo repository.ActivityRepository,
	twitterGW gateway.TwitterGateway,
	jwtSvc *auth.JWTService,
	txManager repository.TransactionManager,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:        userRepo,
		tweetConnRepo:   tweetConnRepo,
		notiSettingRepo: notiSettingRepo,
		activityRepo:    activityRepo,
		twitterGW:       twitterGW,
		jwtSvc:          jwtSvc,
		txManager:       txManager,
	}
}

// XAuth exchanges an OAuth authorization code for tokens, upserts the user,
// saves the Twitter connection, and returns a JWT.
func (u *AuthUsecase) XAuth(ctx context.Context, code string, codeVerifier string) (*XAuthResult, error) {
	// 1. Exchange code for tokens
	tokenResp, err := u.twitterGW.ExchangeCode(ctx, code, codeVerifier)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeUnauthenticated, "failed to exchange OAuth code", err)
	}

	// 2. Get user info from X API
	userInfo, err := u.twitterGW.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to get X user info", err)
	}

	// 3. Upsert user
	user, err := u.userRepo.UpsertByTwitterID(ctx, entity.UpsertUserInput{
		TwitterID:     userInfo.ID,
		Name:          userInfo.Name,
		Email:         userInfo.Email,
		Image:         userInfo.Image,
		TwitterHandle: userInfo.Username,
	})
	if err != nil {
		return nil, err
	}

	// Best effort: set email from X account if user has no email yet
	if userInfo.Email != "" && user.Email == "" {
		if err := u.userRepo.UpdateEmail(ctx, user.ID, userInfo.Email); err != nil {
			slog.Warn("failed to set email from X account", "userID", user.ID, "error", err)
		} else {
			user.Email = userInfo.Email
		}
	}

	tutorialPending := user.IsTutorialPending()

	// 4. Save Twitter connection
	now := time.Now()
	conn := &entity.TwitterConnection{
		UserID:         user.ID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		TokenExpiresAt: now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Status:         entity.TwitterConnected,
		ConnectedAt:    &now,
	}
	if err := u.tweetConnRepo.Upsert(ctx, conn); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to save Twitter connection", err)
	}

	// 5. Best effort: set default notification settings (new users only)
	if tutorialPending {
		defaultSettings := &entity.NotificationSetting{
			UserID:        user.ID,
			SpikeEnabled:  true,
			RisingEnabled: true,
		}
		if err := u.notiSettingRepo.Upsert(ctx, defaultSettings); err != nil {
			slog.Warn("failed to create default notification settings", "userID", user.ID, "error", err)
		}
	}

	// 6. Best effort: record login activity
	RecordActivity(ctx, u.activityRepo, user.ID, entity.ActivityLogin, "", "X アカウントでログインしました")

	// 7. Generate JWT
	token, err := u.jwtSvc.GenerateToken(user.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate JWT", err)
	}

	return &XAuthResult{
		User:            user,
		Token:           token,
		TutorialPending: tutorialPending,
	}, nil
}

// Logout revokes X tokens (best-effort) and removes the user's Twitter connection.
// DeleteByUserID が not_found を返す場合（既に削除済み or DB にレコードなし）は正常終了とする。
func (u *AuthUsecase) Logout(ctx context.Context, userID string) error {
	// Best-effort: revoke X tokens before deleting the connection
	conn, err := u.tweetConnRepo.FindByUserID(ctx, userID)
	if err != nil {
		if !apperror.IsCode(err, apperror.CodeNotFound) {
			slog.Warn("failed to find twitter connection for revoke", "userID", userID, "error", err)
		}
	} else {
		if err := u.twitterGW.RevokeToken(ctx, conn.AccessToken, conn.RefreshToken); err != nil {
			slog.Warn("failed to revoke X tokens", "userID", userID, "error", err)
		}
	}

	if err := u.tweetConnRepo.DeleteByUserID(ctx, userID); err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// GetCurrentUser returns the user, AI generation count, and tutorial status.
func (u *AuthUsecase) GetCurrentUser(ctx context.Context, userID string) (*GetCurrentUserResult, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &GetCurrentUserResult{
		User:            user,
		TutorialPending: user.IsTutorialPending(),
	}, nil
}

// CompleteTutorial marks the user's onboarding tutorial as completed.
func (u *AuthUsecase) CompleteTutorial(ctx context.Context, userID string) error {
	return u.userRepo.CompleteTutorial(ctx, userID)
}

// DeleteAccount deletes the user's data.
func (u *AuthUsecase) DeleteAccount(ctx context.Context, userID string) error {
	return u.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := u.tweetConnRepo.DeleteByUserID(ctx, userID); err != nil {
			return err
		}
		return u.userRepo.Delete(ctx, userID)
	})
}
