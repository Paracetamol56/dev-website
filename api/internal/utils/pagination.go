package utils

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pagination struct {
	Limit int64 `form:"limit,default=10" binding:"gt=0,lte=100"`
	Page  int64 `form:"page,default=1" binding:"gt=0"`
}

func NewPagination(limit int64, page int64) *Pagination {
	return &Pagination{
		Limit: limit,
		Page:  page,
	}
}

func (m *Pagination) GetFindOptions() *options.FindOptions {
	opt := options.FindOptions{Limit: &m.Limit, Skip: &m.Page}
	return &opt
}

func (m *Pagination) GetFacetStage() bson.D {
	return bson.D{
		{"$facet", bson.D{
			{"items", bson.A{
				bson.D{{"$skip", m.Limit * (m.Page - 1)}},
				bson.D{{"$limit", m.Limit}},
			}},
			{"total_items", bson.A{
				bson.D{{"$count", "count"}},
			}},
		}},
	}
}

func (m *Pagination) GetTotalPages(totalItems int64) int64 {
	if m.Limit == 0 {
		return 0
	}
	return (totalItems + m.Limit - 1) / m.Limit
}
