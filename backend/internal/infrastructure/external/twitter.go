package external

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/gateway"
)

const twitterBaseURL = "https://api.x.com"

var _ gateway.TwitterGateway = (*TwitterClient)(nil)

// TwitterClient implements gateway.TwitterGateway using X API v2.
type TwitterClient struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewTwitterClient creates a new X API v2 client.
func NewTwitterClient(clientID, clientSecret, redirectURI string) *TwitterClient {
	return &TwitterClient{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (c *TwitterClient) BuildAuthorizationURL(_ context.Context) (*gateway.OAuthStartResult, error) {
	// code_verifier: 43-128 chars from [A-Za-z0-9-._~]
	verifierBytes := make([]byte, 64)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate code verifier", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// code_challenge: SHA256 + base64url
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// state: random string
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate state", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", c.redirectURI)
	params.Set("scope", "tweet.read tweet.write users.read users.email dm.read dm.write offline.access")
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("prompt", "consent")

	authURL := "https://twitter.com/i/oauth2/authorize?" + params.Encode()

	return &gateway.OAuthStartResult{
		AuthorizationURL: authURL,
		CodeVerifier:     codeVerifier,
		State:            state,
	}, nil
}

func (c *TwitterClient) ExchangeCode(ctx context.Context, code string, codeVerifier string) (*gateway.OAuthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.redirectURI)
	form.Set("code_verifier", codeVerifier)

	body, err := c.doTokenRequest(ctx, form)
	if err != nil {
		return nil, err
	}

	var resp tokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse token response", err)
	}

	return &gateway.OAuthTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

func (c *TwitterClient) RefreshToken(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	body, err := c.doTokenRequest(ctx, form)
	if err != nil {
		return nil, err
	}

	var resp tokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse token response", err)
	}

	return &gateway.OAuthTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

func (c *TwitterClient) GetUserInfo(ctx context.Context, accessToken string) (*gateway.TwitterUserInfo, error) {
	reqURL := twitterBaseURL + "/2/users/me?user.fields=id,name,username,profile_image_url,confirmed_email"

	body, err := c.doRequest(ctx, http.MethodGet, reqURL, nil, "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
			Email           string `json:"confirmed_email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse user info response", err)
	}

	return &gateway.TwitterUserInfo{
		ID:       resp.Data.ID,
		Name:     resp.Data.Name,
		Username: resp.Data.Username,
		Email:    resp.Data.Email,
		Image:    resp.Data.ProfileImageURL,
	}, nil
}

func (c *TwitterClient) SearchRecentTweets(ctx context.Context, accessToken string, input gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
	params := url.Values{}
	params.Set("query", input.Query)
	params.Set("tweet.fields", "id,text,author_id,conversation_id,created_at,public_metrics")
	params.Set("expansions", "author_id")
	params.Set("user.fields", "name,username")
	if input.MaxResults > 0 {
		params.Set("max_results", fmt.Sprintf("%d", input.MaxResults))
	}
	if input.StartTime != nil {
		params.Set("start_time", input.StartTime.Format(time.RFC3339))
	}
	if input.EndTime != nil {
		params.Set("end_time", input.EndTime.Format(time.RFC3339))
	}
	if input.SinceID != nil && *input.SinceID != "" {
		params.Set("since_id", *input.SinceID)
	}
	if input.SortOrder != "" {
		params.Set("sort_order", input.SortOrder)
	}

	reqURL := twitterBaseURL + "/2/tweets/search/recent?" + params.Encode()

	body, err := c.doRequest(ctx, http.MethodGet, reqURL, nil, "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			ID             string `json:"id"`
			Text           string `json:"text"`
			AuthorID       string `json:"author_id"`
			ConversationID string `json:"conversation_id"`
			CreatedAt      string `json:"created_at"`
			PublicMetrics  struct {
				RetweetCount    int `json:"retweet_count"`
				ReplyCount      int `json:"reply_count"`
				LikeCount       int `json:"like_count"`
				QuoteCount      int `json:"quote_count"`
				ImpressionCount int `json:"impression_count"`
			} `json:"public_metrics"`
		} `json:"data"`
		Includes struct {
			Users []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"users"`
		} `json:"includes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse search response", err)
	}

	// Build author lookup map from includes
	authorMap := make(map[string]struct{ Name, Username string }, len(resp.Includes.Users))
	for _, u := range resp.Includes.Users {
		authorMap[u.ID] = struct{ Name, Username string }{u.Name, u.Username}
	}

	tweets := make([]gateway.Tweet, 0, len(resp.Data))
	for _, d := range resp.Data {
		createdAt, _ := time.Parse(time.RFC3339, d.CreatedAt)
		t := gateway.Tweet{
			ID:             d.ID,
			Text:           d.Text,
			AuthorID:       d.AuthorID,
			ConversationID: d.ConversationID,
			CreatedAt:      createdAt,
			Metrics: gateway.TweetMetrics{
				RetweetCount:    d.PublicMetrics.RetweetCount,
				ReplyCount:      d.PublicMetrics.ReplyCount,
				LikeCount:       d.PublicMetrics.LikeCount,
				QuoteCount:      d.PublicMetrics.QuoteCount,
				ImpressionCount: d.PublicMetrics.ImpressionCount,
			},
		}
		if author, ok := authorMap[d.AuthorID]; ok {
			t.AuthorName = author.Name
			t.AuthorHandle = author.Username
		}
		tweets = append(tweets, t)
	}

	return tweets, nil
}

func (c *TwitterClient) GetTweetCounts(ctx context.Context, accessToken string, query string, startTime time.Time) ([]gateway.TweetCountDataPoint, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start_time", startTime.Format(time.RFC3339))
	params.Set("granularity", "hour")

	reqURL := twitterBaseURL + "/2/tweets/counts/recent?" + params.Encode()

	body, err := c.doRequest(ctx, http.MethodGet, reqURL, nil, "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			Start      string `json:"start"`
			End        string `json:"end"`
			TweetCount int    `json:"tweet_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse tweet counts response", err)
	}

	points := make([]gateway.TweetCountDataPoint, 0, len(resp.Data))
	for _, d := range resp.Data {
		start, _ := time.Parse(time.RFC3339, d.Start)
		end, _ := time.Parse(time.RFC3339, d.End)
		points = append(points, gateway.TweetCountDataPoint{
			Start:      start,
			End:        end,
			TweetCount: d.TweetCount,
		})
	}

	return points, nil
}

func (c *TwitterClient) PostTweet(ctx context.Context, accessToken string, text string) (string, error) {
	payload := map[string]string{"text": text}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to marshal tweet body", err)
	}

	body, err := c.doRequest(ctx, http.MethodPost, twitterBaseURL+"/2/tweets", jsonBody, "Bearer "+accessToken)
	if err != nil {
		return "", err
	}

	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to parse post tweet response", err)
	}

	tweetURL := fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID)
	return tweetURL, nil
}

func (c *TwitterClient) PostReply(ctx context.Context, accessToken string, text string, inReplyToTweetID string) (string, error) {
	payload := map[string]any{
		"text": text,
		"reply": map[string]string{
			"in_reply_to_tweet_id": inReplyToTweetID,
		},
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to marshal reply body", err)
	}

	body, err := c.doRequest(ctx, http.MethodPost, twitterBaseURL+"/2/tweets", jsonBody, "Bearer "+accessToken)
	if err != nil {
		return "", err
	}

	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to parse post reply response", err)
	}

	return resp.Data.ID, nil
}

func (c *TwitterClient) DeleteTweet(ctx context.Context, accessToken string, tweetID string) error {
	reqURL := fmt.Sprintf("%s/2/tweets/%s", twitterBaseURL, tweetID)

	_, err := c.doRequest(ctx, http.MethodDelete, reqURL, nil, "Bearer "+accessToken)
	return err
}

func (c *TwitterClient) VerifyCredentials(ctx context.Context, accessToken string) error {
	reqURL := twitterBaseURL + "/2/users/me"

	_, err := c.doRequest(ctx, http.MethodGet, reqURL, nil, "Bearer "+accessToken)
	return err
}

func (c *TwitterClient) SendDirectMessage(ctx context.Context, accessToken string, recipientID string, text string) error {
	payload := map[string]string{"text": text}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to marshal dm body", err)
	}

	reqURL := fmt.Sprintf("%s/2/dm_conversations/with/%s/messages", twitterBaseURL, recipientID)

	_, err = c.doRequest(ctx, http.MethodPost, reqURL, jsonBody, "Bearer "+accessToken)
	return err
}

func (c *TwitterClient) RevokeToken(ctx context.Context, accessToken string, refreshToken string) error {
	revoke := func(token, hint string) {
		form := url.Values{}
		form.Set("token", token)
		form.Set("token_type_hint", hint)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, twitterBaseURL+"/2/oauth2/revoke", strings.NewReader(form.Encode()))
		if err != nil {
			slog.Warn("failed to create revoke request", "hint", hint, "error", err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(c.clientID, c.clientSecret)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			slog.Warn("failed to revoke X token", "hint", hint, "error", err)
			return
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)

		if resp.StatusCode >= 300 {
			slog.Warn("X token revoke returned non-success", "hint", hint, "status", resp.StatusCode)
		}
	}

	slog.Info("revoking X tokens", "has_access_token", accessToken != "", "has_refresh_token", refreshToken != "")

	if accessToken != "" {
		revoke(accessToken, "access_token")
	}
	if refreshToken != "" {
		revoke(refreshToken, "refresh_token")
	}

	return nil
}

// --- internal helpers ---

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// doTokenRequest sends an OAuth token request with Basic Auth.
func (c *TwitterClient) doTokenRequest(ctx context.Context, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twitterBaseURL+"/2/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create token request", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "twitter token request failed", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to read token response body", err)
	}

	if err := checkTwitterResponse(resp.StatusCode, body); err != nil {
		return nil, err
	}

	return body, nil
}

// doRequest performs an authenticated HTTP request to the X API.
func (c *TwitterClient) doRequest(ctx context.Context, method, reqURL string, jsonBody []byte, authHeader string) ([]byte, error) {
	var bodyReader io.Reader
	if jsonBody != nil {
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create request", err)
	}
	req.Header.Set("Authorization", authHeader)
	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "twitter API request failed", err)
	}
	defer resp.Body.Close()

	if remaining := resp.Header.Get("x-rate-limit-remaining"); remaining != "" {
		if n, err := strconv.Atoi(remaining); err == nil && n <= 10 {
			slog.Warn("twitter rate limit low", "remaining", n, "endpoint", reqURL)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to read response body", err)
	}

	if err := checkTwitterResponse(resp.StatusCode, body); err != nil {
		return nil, err
	}

	return body, nil
}

// checkTwitterResponse maps HTTP status codes to apperror codes.
func checkTwitterResponse(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	detail := string(body)
	switch {
	case statusCode == 400:
		return apperror.Wrap(apperror.CodeInvalidArgument, "twitter: bad request", fmt.Errorf("%s", detail))
	case statusCode == 401:
		return apperror.Wrap(apperror.CodeUnauthenticated, "twitter: unauthorized", fmt.Errorf("%s", detail))
	case statusCode == 403:
		return apperror.Wrap(apperror.CodePermissionDenied, "twitter: forbidden", fmt.Errorf("%s", detail))
	case statusCode == 404:
		return apperror.Wrap(apperror.CodeNotFound, "twitter: not found", fmt.Errorf("%s", detail))
	case statusCode == 429:
		return apperror.Wrap(apperror.CodeResourceExhausted, "twitter: rate limit exceeded", fmt.Errorf("%s", detail))
	default:
		return apperror.Wrap(apperror.CodeInternal, fmt.Sprintf("twitter: unexpected status %d", statusCode), fmt.Errorf("%s", detail))
	}
}
