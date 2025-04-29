package utils

type PaginatedArrayResult[T any] struct {
	Page       int64 `json:"page"`
	Limit      int64 `json:"limit"`
	TotalPages int64 `json:"total_pages"`
	TotalItems int64 `json:"total_items" bson:"total_items"`
	Items      []*T  `json:"items" bson:"items"`
}

type ArrayResult[T any] struct {
	TotalItems int64 `json:"total_items"`
	Items      []*T  `json:"items"`
}

type countMongoResult struct {
	Count int64 `bson:"count"`
}

type PaginatedFacetResult[T any] struct {
	Items      []*T               `bson:"items"`
	TotalItems []countMongoResult `bson:"total_items"`
}

func (r *PaginatedFacetResult[T]) getCount() int64 {
	if len(r.TotalItems) == 0 {
		return 0
	}
	return r.TotalItems[0].Count
}

func (r *PaginatedFacetResult[T]) ToPaginatedArrayResult(pagination *Pagination) *PaginatedArrayResult[T] {
	return &PaginatedArrayResult[T]{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: pagination.GetTotalPages(r.getCount()),
		TotalItems: r.getCount(),
		Items:      r.Items,
	}
}

func (r *PaginatedFacetResult[T]) ToArrayResult() *ArrayResult[T] {
	return &ArrayResult[T]{
		TotalItems: r.getCount(),
		Items:      r.Items,
	}
}
