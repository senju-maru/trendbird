package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
)

const claudeBaseURL = "https://api.anthropic.com/v1"

var _ gateway.AIGateway = (*ClaudeClient)(nil)

// ClaudeClient implements gateway.AIGateway using the Anthropic Messages API.
type ClaudeClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

// NewClaudeClient creates a new Claude API client.
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		apiKey:     apiKey,
		model:      "claude-haiku-4-5-20251001",
	}
}

// ResearchTopic performs a web search via Claude's web_search tool and returns a summary.
func (c *ClaudeClient) ResearchTopic(ctx context.Context, input gateway.ResearchTopicInput) (*gateway.ResearchTopicOutput, error) {
	keywordsText := strings.Join(input.Keywords, ", ")
	userPrompt := fmt.Sprintf(`以下のトピックについてWeb検索を行い、最新の情報を日本語で要約してください。

トピック名: %s
関連キーワード: %s

要求:
- 最新のニュース・動向を中心に要約してください
- 情報源のURLを可能な限り収集してください
- 要約は300〜500文字程度にしてください
- 必ず以下のJSON形式で返してください:
{"summary": "要約テキスト", "source_urls": ["url1", "url2", ...]}`,
		input.TopicName,
		keywordsText,
	)

	reqBody := claudeMessagesRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Tools: []claudeTool{
			{
				Type: "web_search_20250305",
				Name: "web_search",
			},
		},
		Messages: []claudeMessage{
			textMessage("user", userPrompt),
		},
	}

	content, err := c.sendMessages(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	var result struct {
		Summary    string   `json:"summary"`
		SourceURLs []string `json:"source_urls"`
	}
	jsonStr := extractJSON(content)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		slog.Error("failed to parse Claude research response", "body", content, "error", err)
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse Claude research response", err)
	}

	return &gateway.ResearchTopicOutput{
		Summary:    result.Summary,
		SourceURLs: result.SourceURLs,
	}, nil
}

// GeneratePosts generates X post content using Claude.
func (c *ClaudeClient) GeneratePosts(ctx context.Context, input gateway.GeneratePostsInput) (*gateway.GeneratePostsOutput, error) {
	styleInstruction := claudeStyleToInstruction(input.Style)

	var researchSection string
	if len(input.TopicContext.ResearchResults) > 0 {
		researchSection = "\n\n【最新の調査結果】\n"
		for i, r := range input.TopicContext.ResearchResults {
			researchSection += fmt.Sprintf("%d. %s\n", i+1, r)
		}
	}

	var keywordsText string
	if len(input.TopicContext.TrendKeywords) > 0 {
		keywordsText = fmt.Sprintf("\nトレンドキーワード: %s", strings.Join(input.TopicContext.TrendKeywords, ", "))
	}

	systemPrompt := `あなたはX(Twitter)の投稿文を生成するAIアシスタントです。
指定されたトピック・文脈・スタイルに基づいて、魅力的なX投稿文を生成してください。
各投稿は140文字以内で、自然な日本語で書いてください。
最新の調査結果が提供されている場合は、それを踏まえてタイムリーな投稿を作成してください。
必ず以下のJSON形式で返してください:
{"posts": ["投稿文1", "投稿文2", ...]}`

	userPrompt := fmt.Sprintf(`トピック: %s
概要: %s%s%s
スタイル: %s
生成数: %d`,
		input.TopicName,
		input.TopicContext.Summary,
		keywordsText,
		researchSection,
		styleInstruction,
		input.Count,
	)

	reqBody := claudeMessagesRequest{
		Model:     c.model,
		MaxTokens: 1500,
		System:    systemPrompt,
		Messages: []claudeMessage{
			textMessage("user", userPrompt),
		},
	}

	content, err := c.sendMessages(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	var result struct {
		Posts []string `json:"posts"`
	}
	jsonStr := extractJSON(content)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		slog.Error("failed to parse Claude generate response", "body", content, "error", err)
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to parse Claude response for post generation", err)
	}

	return &gateway.GeneratePostsOutput{Posts: result.Posts}, nil
}

// --- internal helpers ---

type claudeMessagesRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Tools     []claudeTool    `json:"tools,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeTool struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func textMessage(role, text string) claudeMessage {
	b, _ := json.Marshal(text)
	return claudeMessage{Role: role, Content: b}
}

type claudeMessagesResponse struct {
	Content    []claudeContentBlock `json:"content"`
	StopReason string               `json:"stop_reason"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// sendMessages sends a request to the Claude Messages API and returns the text content.
// For web_search tool requests, it handles the agentic loop (tool_use → continue until end_turn).
func (c *ClaudeClient) sendMessages(ctx context.Context, reqBody claudeMessagesRequest) (string, error) {
	// For requests with tools, we need an agentic loop.
	// Claude will call tools, we pass back the full response, until stop_reason is end_turn.
	messages := reqBody.Messages
	const maxIterations = 10

	for range maxIterations {
		reqBody.Messages = messages

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return "", apperror.Wrap(apperror.CodeInternal, "failed to marshal Claude request", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeBaseURL+"/messages", bytes.NewReader(jsonBody))
		if err != nil {
			return "", apperror.Wrap(apperror.CodeInternal, "failed to create Claude request", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", apperror.Wrap(apperror.CodeInternal, "Claude API request failed", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", apperror.Wrap(apperror.CodeInternal, "failed to read Claude response body", err)
		}

		if err := checkClaudeResponse(resp.StatusCode, respBody); err != nil {
			return "", err
		}

		var chatResp claudeMessagesResponse
		if err := json.Unmarshal(respBody, &chatResp); err != nil {
			return "", apperror.Wrap(apperror.CodeInternal, "failed to parse Claude response", err)
		}

		// If stop_reason is end_turn or no tools, extract text and return
		if chatResp.StopReason == "end_turn" || len(reqBody.Tools) == 0 {
			return extractTextContent(chatResp.Content), nil
		}

		// For tool_use / pause_turn, re-send the conversation with the raw response appended.
		// The Messages API handles server-side tools internally, so we just need to pass the
		// assistant response back and let the API continue.
		if chatResp.StopReason == "tool_use" || chatResp.StopReason == "pause_turn" {
			// Rebuild messages: original user message + assistant response as raw JSON
			messages = []claudeMessage{
				reqBody.Messages[0],
			}
			// Append the assistant response content array as-is for the next iteration
			assistantContent, _ := json.Marshal(chatResp.Content)
			messages = append(messages, claudeMessage{
				Role:    "assistant",
				Content: assistantContent,
			})
			continue
		}

		// Unknown stop_reason — extract what we have
		return extractTextContent(chatResp.Content), nil
	}

	return "", apperror.Internal("Claude API: max iterations reached")
}

func extractTextContent(blocks []claudeContentBlock) string {
	// Return the LAST text block. For web_search responses, the first text block
	// is often a preamble ("検索させていただきます") while the actual result is in the last one.
	var lastText string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			lastText = block.Text
		}
	}
	return lastText
}

// extractJSON finds and extracts a JSON object from text that may contain surrounding prose.
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start == -1 {
		return text
	}
	// Find the matching closing brace
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return text
}

func claudeStyleToInstruction(style entity.PostStyle) string {
	switch style {
	case entity.PostStyleCasual:
		return "カジュアルで親しみやすい口調"
	case entity.PostStyleBreaking:
		return "ニュース速報風の緊迫感ある口調"
	case entity.PostStyleAnalysis:
		return "データ・事実に基づく分析口調"
	default:
		return "カジュアルで親しみやすい口調"
	}
}

func checkClaudeResponse(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	detail := string(body)
	switch statusCode {
	case 400:
		return apperror.Wrap(apperror.CodeInvalidArgument, "claude: bad request", fmt.Errorf("%s", detail))
	case 401:
		return apperror.Wrap(apperror.CodeInternal, "claude: invalid API key", fmt.Errorf("%s", detail))
	case 429:
		return apperror.Wrap(apperror.CodeResourceExhausted, "claude: rate limit exceeded", fmt.Errorf("%s", detail))
	case 529:
		return apperror.Wrap(apperror.CodeResourceExhausted, "claude: API overloaded", fmt.Errorf("%s", detail))
	default:
		return apperror.Wrap(apperror.CodeInternal, fmt.Sprintf("claude: unexpected status %d", statusCode), fmt.Errorf("%s", detail))
	}
}
