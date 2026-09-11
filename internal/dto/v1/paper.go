package apiv1

type PaperQuery struct {
	PageQuery
	Type    string `form:"type" json:"type"`
	Year    int    `form:"year" json:"year"`
	Include string `form:"include" json:"include"`
}
