package shared

type PaginationParams struct {
	Page    int `query:"page" validate:"min=1"`
	PerPage int `query:"per_page" validate:"min=1,max=100"`
}

func (p *PaginationParams) Defaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PerPage == 0 {
		p.PerPage = 20
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type PaginatedResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PerPage  int `json:"per_page"`
	LastPage int `json:"last_page"`
}

func NewPaginatedResult[T any](items []T, total, page, perPage int) PaginatedResult[T] {
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}
	return PaginatedResult[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
		LastPage: lastPage,
	}
}
