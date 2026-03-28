package converter

import (
	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// GenreToProto converts a domain Genre entity to a Proto Genre message.
func GenreToProto(e *entity.Genre) *trendbirdv1.Genre {
	return &trendbirdv1.Genre{
		Id:          e.ID,
		Slug:        e.Slug,
		Label:       e.Label,
		Description: e.Description,
		SortOrder:   int32(e.SortOrder),
	}
}

// GenreSliceToProto converts a slice of Genre entities to Proto messages.
func GenreSliceToProto(es []*entity.Genre) []*trendbirdv1.Genre {
	result := make([]*trendbirdv1.Genre, len(es))
	for i, e := range es {
		result[i] = GenreToProto(e)
	}
	return result
}
