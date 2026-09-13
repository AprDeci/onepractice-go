package apiv1

type CreateRecordRequest struct {
	PaperID      int    `json:"paperId" binding:"required"`
	Type         string `json:"type"`
	IsFinished   int    `json:"isFinished"`
	Answers      string `json:"answers"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"totalScore"`
	TimeSpend    int    `json:"timeSpend"`
	HasSpendTime int64  `json:"hasSpendTime"`
}

type UpdateRecordRequest struct {
	PaperID      int    `json:"paperId" binding:"required"`
	Type         string `json:"type"`
	IsFinished   int    `json:"isFinished"`
	Answers      string `json:"answers"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"totalScore"`
	TimeSpend    int    `json:"timeSpend"`
	HasSpendTime int64  `json:"hasSpendTime"`
}

type RecordListQuery struct {
	PageQuery
	Days int `form:"days" json:"days"`
}
