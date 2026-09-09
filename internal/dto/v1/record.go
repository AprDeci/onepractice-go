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

type RecordCreatedResponse struct {
	RecordID string `json:"recordId"`
}

type UserExamRecord struct {
	RecordID     string `json:"recordId"`
	UserID       int64  `json:"userId"`
	PaperID      int    `json:"paperId"`
	PaperType    string `json:"paperType"`
	PaperName    string `json:"paperName"`
	Type         string `json:"type"`
	IsFinished   int    `json:"isFinished"`
	Answers      string `json:"answers"`
	TimeSpend    int    `json:"timeSpend"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"totalScore"`
	Timestamp    int64  `json:"timestamp"`
	HasSpendTime int64  `json:"hasSpendTime"`
}

type RecordListQuery struct {
	PageQuery
	Days int `form:"days" json:"days"`
}
