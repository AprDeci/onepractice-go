package apiv1

type PageQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

func (q PageQuery) Normalize() PageQuery {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q
}

func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}
