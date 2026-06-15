package dto

type RecordRequest struct {
	RecordID     string `json:"recordId"`
	PaperID      int    `json:"paperId" binding:"required"`
	Type         string `json:"type"`
	IsFinished   int    `json:"isfinished"`
	Answers      string `json:"answers"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"totalscore"`
	TimeSpend    int    `json:"timespend"`
	HasSpendTime int64  `json:"hasspendtime"`
}

type UserExamRecord struct {
	RecordID     string `json:"recordId"`
	UserID       int64  `json:"userId"`
	PaperID      int    `json:"paperId"`
	PaperType    string `json:"paperType"`
	PaperName    string `json:"paperName"`
	Type         string `json:"type"`
	IsFinished   int    `json:"isfinished"`
	Answers      string `json:"answers"`
	TimeSpend    int    `json:"timespend"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"totalscore"`
	Timestamp    int64  `json:"timestamp"`
	HasSpendTime int64  `json:"hasspendtime"`
}
