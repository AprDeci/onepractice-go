package model

type Paper struct {
	PaperID       int    `gorm:"column:paper_id;primaryKey" json:"paperId"`
	PaperName     string `gorm:"column:paper_name" json:"paperName"`
	ExamYear      int    `gorm:"column:exam_year" json:"examYear"`
	ExamMonth     int    `gorm:"column:exam_month" json:"examMonth"`
	Version       int    `gorm:"column:version" json:"version"`
	TotalTime     int    `gorm:"column:total_time" json:"totalTime"`
	Type          string `gorm:"column:type" json:"type"`
	QuestionCount int64  `gorm:"-" json:"questionCount,omitempty"`
}

func (Paper) TableName() string { return "papers" }

type PaperRateMapping struct {
	PaperID int `gorm:"column:paperId;primaryKey" json:"paperid"`
	Rating  int `gorm:"column:rating" json:"rating"`
	Number  int `gorm:"column:number" json:"number"`
}

func (PaperRateMapping) TableName() string { return "paper_rate_mapping" }
