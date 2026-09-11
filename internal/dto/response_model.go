package dto

import (
	"time"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/model"
)

// 通用分页

type PageResult[T any] struct {
	Total int64 `json:"total"`
	Data  []T   `json:"data"`
}

type PageListResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// 分页数据的具体类型别名，供 swagger 注解引用（swag 无法直接解析泛型实例）。
type (
	DictionaryWordListPage = PageListResult[DictionaryWordListItem]
	DictionaryBookListPage = PageListResult[DictionaryBookListItem]
	PaperPage              = PageResult[model.Paper]
	PaperWithRatingPage    = PageResult[PaperWithRating]
)

// 认证

type ResetPasswordTokenResponse struct {
	ResetToken string `json:"resetToken"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
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

// 词典

type DictionaryWordListItem struct {
	WordID     uint    `gorm:"column:wordid" json:"wordid"`
	Spelling   string  `gorm:"column:spelling" json:"spelling"`
	UKPhonetic string  `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	USPhonetic string  `gorm:"column:us_phonetic" json:"us_phonetic"`
	Paraphrase string  `gorm:"column:paraphrase" json:"paraphrase"`
	Frequency  float64 `gorm:"column:frequency" json:"frequency"`
}

type DictionaryWordExampleItem struct {
	ExaPID  int    `gorm:"column:exapid" json:"exapid"`
	EN      string `gorm:"column:en" json:"en"`
	CN      string `gorm:"column:cn" json:"cn"`
	Heat    *int   `gorm:"column:heat" json:"heat"`
	AddDate string `gorm:"column:adddate" json:"adddate"`
}

type DictionaryBookSimple struct {
	BookID   uint   `gorm:"column:bookid" json:"bookid"`
	BookName string `gorm:"column:bookname" json:"bookname"`
}

type DictionaryWordDetail struct {
	Word     DictionaryWordListItem      `json:"word"`
	Books    []DictionaryBookSimple      `json:"books"`
	Examples []DictionaryWordExampleItem `json:"examples"`
}

type DictionaryBookListItem struct {
	BookID   uint   `gorm:"column:bookid" json:"bookid"`
	BookName string `gorm:"column:bookname" json:"bookname"`
	VocCount *int   `gorm:"column:voccount" json:"voccount"`
	Status   *int   `gorm:"column:status" json:"status"`
}

type DictionaryLookupResult struct {
	Spelling string                   `json:"spelling"`
	Exact    bool                     `json:"exact"`
	Total    int                      `json:"total"`
	Items    []DictionaryWordListItem `json:"items"`
}

// 作文

// EssayTask 是作文批改异步任务的状态与结果。
type EssayTask struct {
	ID        string        `json:"taskId"`
	UserID    int64         `json:"userId"`
	RecordID  string        `json:"recordId,omitempty"`
	Status    string        `json:"status"`
	Input     agent.Input   `json:"input"`
	Result    *agent.Output `json:"result,omitempty"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// EssayResultResponse 是作文评分结果的对外返回结构。
type EssayResultResponse struct {
	TaskID         string        `json:"taskId"`
	RecordID       string        `json:"recordId,omitempty"`
	Title          string        `json:"title"`
	FullScore      int           `json:"fullScore"`
	TotalScore     float64       `json:"totalScore"`
	GrammarScore   float64       `json:"grammarScore"`
	TopicScore     float64       `json:"topicScore"`
	WordScore      float64       `json:"wordScore"`
	StructureScore float64       `json:"structureScore"`
	WordNum        int           `json:"wordNum"`
	Result         *agent.Output `json:"result,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
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

// 收藏

type CollectedWordItem struct {
	ID         uint      `gorm:"column:id" json:"id"`
	FavoriteID uint64    `gorm:"column:favorite_id" json:"favorite_id"`
	WordID     uint      `gorm:"column:wordid" json:"wordid"`
	Word       string    `gorm:"column:word" json:"word"`
	Spelling   string    `gorm:"column:spelling" json:"spelling"`
	UKPhonetic string    `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	USPhonetic string    `gorm:"column:us_phonetic" json:"us_phonetic"`
	Paraphrase string    `gorm:"column:paraphrase" json:"paraphrase"`
	Frequency  float64   `gorm:"column:frequency" json:"frequency"`
	PaperID    *int      `gorm:"column:paper_id" json:"paper_id"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

type CollectedWordList struct {
	Total int64               `json:"total"`
	Data  []CollectedWordItem `json:"data"`
}

// 健康检查

// HealthResponse 是健康检查接口返回的数据结构。
type HealthResponse struct {
	Status   string `json:"status"`
	Database bool   `json:"database"`
}
