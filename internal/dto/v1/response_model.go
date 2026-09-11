package apiv1

import (
	"time"

	"onepractice-golang/internal/model"
)

// 通用分页

type Page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// 分页数据的具体类型别名，供 swagger 注解引用（swag 无法直接解析泛型实例）。
type (
	DictionaryWordPage = Page[DictionaryWordListItem]
	DictionaryBookPage = Page[DictionaryBookListItem]
	PaperPage          = Page[model.Paper]
	RecordPage         = Page[UserExamRecord]
	FavoriteWordPage   = Page[CollectedWordItem]
)

// 认证

type RegisterResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type UserInfoResponse struct {
	Username string `json:"username"`
	UserType int    `json:"userType"`
	Email    string `json:"email"`
}

type ResetTokenResponse struct {
	ResetToken string `json:"resetToken"`
}

// 词典

type DictionaryWordListItem struct {
	WordID     uint    `json:"wordId"`
	Spelling   string  `json:"spelling"`
	UKPhonetic string  `json:"ukPhonetic"`
	USPhonetic string  `json:"usPhonetic"`
	Paraphrase string  `json:"paraphrase"`
	Frequency  float64 `json:"frequency"`
}

type DictionaryWordExampleItem struct {
	ExaPID  int    `json:"exaPid"`
	EN      string `json:"en"`
	CN      string `json:"cn"`
	Heat    *int   `json:"heat"`
	AddDate string `json:"addDate"`
}

type DictionaryBookSimple struct {
	BookID   uint   `json:"bookId"`
	BookName string `json:"bookName"`
}

type DictionaryWordDetail struct {
	Word     DictionaryWordListItem      `json:"word"`
	Books    []DictionaryBookSimple      `json:"books"`
	Examples []DictionaryWordExampleItem `json:"examples"`
}

type DictionaryBookListItem struct {
	BookID   uint   `json:"bookId"`
	BookName string `json:"bookName"`
	VocCount *int   `json:"vocCount"`
	Status   *int   `json:"status"`
}

type DictionaryLookupResult struct {
	Spelling string                   `json:"spelling"`
	Exact    bool                     `json:"exact"`
	Total    int                      `json:"total"`
	Items    []DictionaryWordListItem `json:"items"`
}

// 试卷

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

// 题目

type QuestionPart struct {
	Questions []model.Question `json:"questions"`
}

type ExamQuestion struct {
	PaperID       int            `json:"paperId"`
	QuestionParts []QuestionPart `json:"questionParts"`
}

type AnswersResponse struct {
	PaperID int           `json:"paperId"`
	Answers model.Answers `json:"answers"`
}

type PracticeQuestionGroup struct {
	PaperID       int            `json:"paperId"`
	QuestionParts []QuestionPart `json:"questionParts"`
	Answers       model.Answers  `json:"answers"`
}

type PracticeQuestionResponse struct {
	QuestionType string                  `json:"questionType"`
	Groups       []PracticeQuestionGroup `json:"groups"`
}

// 记录

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

// 收藏

type FavoriteStatusResponse struct {
	WordID    uint `json:"wordId"`
	Favorited bool `json:"favorited"`
}

type CollectedWordItem struct {
	ID         uint      `json:"id"`
	FavoriteID uint64    `json:"favoriteId"`
	WordID     uint      `json:"wordId"`
	Word       string    `json:"word"`
	Spelling   string    `json:"spelling"`
	UKPhonetic string    `json:"ukPhonetic"`
	USPhonetic string    `json:"usPhonetic"`
	Paraphrase string    `json:"paraphrase"`
	Frequency  float64   `json:"frequency"`
	PaperID    *int      `json:"paperId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CollectedWordList struct {
	Total int64               `json:"total"`
	Data  []CollectedWordItem `json:"data"`
}

// 作文

// EssayTaskCreatedResponse 是创建作文批改任务后返回的数据结构。
type EssayTaskCreatedResponse struct {
	TaskID string `json:"taskId"`
	Status string `json:"status"`
}
