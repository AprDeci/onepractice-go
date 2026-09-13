package dto

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

func (r *PageQuery) Normalize() {
	if r.Page <= 0 {
		r.Page = DefaultPage
	}
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	}
	if r.PageSize > MaxPageSize {
		r.PageSize = MaxPageSize
	}
}

func (r PageQuery) Offset() int {
	return (r.Page - 1) * r.PageSize
}
