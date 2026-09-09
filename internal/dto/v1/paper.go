package apiv1

type PaperQuery struct {
	PageQuery
	Type    string `form:"type" json:"type"`
	Year    int    `form:"year" json:"year"`
	Include string `form:"include" json:"include"`
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
