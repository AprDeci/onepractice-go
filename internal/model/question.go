package model

type Question struct {
	QuestionID           int        `gorm:"column:question_id;primaryKey" json:"questionId"`
	PaperID              int        `gorm:"column:paper_id" json:"paperId"`
	PartName             string     `gorm:"column:part_name" json:"partName"`
	SectionName          string     `gorm:"column:section_name" json:"sectionName"`
	QuestionType         string     `gorm:"column:question_type" json:"questionType"`
	QuestionOrder        int        `gorm:"column:question_order" json:"questionOrder"`
	Content              string     `gorm:"column:content" json:"content"`
	CorrectAnswer        Answers    `gorm:"column:correct_answer;type:json" json:"correctAnswer"`
	ReadingSplitQuestion StringList `gorm:"column:reading_split_question;type:json" json:"readingSplitQuestion"`
	Options              Options    `gorm:"column:options;type:json" json:"options"`
	WordBank             StringList `gorm:"column:word_bank;type:json" json:"wordBank"`
	MatchingData         ReadItems  `gorm:"column:matching_data;type:json" json:"matchingData"`
	ListenURL            string     `gorm:"column:listenurl" json:"listenurl"`
}

func (Question) TableName() string { return "questions" }
