package gateway

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// GeneratePostsInput holds the parameters for AI post generation.
type GeneratePostsInput struct {
	TopicName    string
	TopicContext TopicContext
	Style        entity.PostStyle
	Count        int
}

// TopicContext provides contextual information about a topic for AI analysis.
type TopicContext struct {
	TrendKeywords   []string
	Summary         string
	ResearchResults []string
}

// GeneratePostsOutput holds the results of AI post generation.
type GeneratePostsOutput struct {
	Posts []string
}

// ResearchTopicInput holds the parameters for web-based topic research.
type ResearchTopicInput struct {
	TopicName string
	Keywords  []string
}

// ResearchTopicOutput holds the results of a web search topic research.
type ResearchTopicOutput struct {
	Summary    string
	SourceURLs []string
}

// AIGateway abstracts the AI API for post generation and topic research.
type AIGateway interface {
	GeneratePosts(ctx context.Context, input GeneratePostsInput) (*GeneratePostsOutput, error)
	ResearchTopic(ctx context.Context, input ResearchTopicInput) (*ResearchTopicOutput, error)
}
