package dto

type PaperQueryRequest struct {
	Page int    `json:"page" binding:"required,min=1"`
	Size int    `json:"size" binding:"required,min=1,max=100"`
	Type string `json:"type"`
	Year int    `json:"year"`
}

type PaperWithRating struct {
	PaperID       int    `json:"paperId"`
	PaperName     string `json:"paperName"`
	ExamYear      int    `json:"examYear"`
	ExamMonth     int    `json:"examMonth"`
	Version       int    `json:"version"`
	TotalTime     int    `json:"totalTime"`
	Type          string `json:"type"`
	QuestionCount int64  `json:"questionCount"`
	Rating        int    `json:"rating"`
	Number        int    `json:"number"`
}

type PaperIntro struct {
	PaperName            string `json:"paperName"`
	ExamYear             int    `json:"examYear"`
	ExamMonth            int    `json:"examMonth"`
	PaperType            string `json:"paperType"`
	PaperTime            int    `json:"paperTime"`
	Difficulty           string `json:"difficulty"`
	SectionCount         int64  `json:"sectionCount"`
	SectionQuestionCount []int  `json:"sectionQuestionCount"`
}
