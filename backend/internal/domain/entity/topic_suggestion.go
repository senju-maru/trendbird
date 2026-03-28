package entity

// TopicSuggestion represents a fuzzy-matched topic candidate from pg_trgm search.
type TopicSuggestion struct {
	ID              string
	Name            string
	Keywords        []string
	GenreSlug       string
	GenreLabel      string
	SimilarityScore float64
}
