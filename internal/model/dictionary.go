package model

type Vocabulary struct {
	WordID     uint    `gorm:"column:wordid;primaryKey;autoIncrement" json:"wordid"`
	Spelling   string  `gorm:"column:spelling" json:"spelling"`
	UKPhonetic string  `gorm:"column:UKphonetic" json:"uk_phonetic"`
	USPhonetic string  `gorm:"column:USphonetic" json:"us_phonetic"`
	Paraphrase string  `gorm:"column:paraphrase" json:"paraphrase"`
	Frequency  float64 `gorm:"column:frequency" json:"frequency"`
}

func (Vocabulary) TableName() string { return "tb_vocabulary" }

type Book struct {
	BookID   uint   `gorm:"column:bookid;primaryKey;autoIncrement" json:"bookid"`
	BookName string `gorm:"column:bookname" json:"bookname"`
	VocCount *int   `gorm:"column:voccount" json:"voccount"`
	Status   *int   `gorm:"column:status" json:"status"`
}

func (Book) TableName() string { return "tb_book" }

type VocBook struct {
	VocBkID int   `gorm:"column:vocbkid;primaryKey;autoIncrement" json:"vocbkid"`
	WordID  *uint `gorm:"column:wordid" json:"wordid"`
	BookID  *uint `gorm:"column:bookid" json:"bookid"`
}

func (VocBook) TableName() string { return "tb_voc_book" }

type VocExample struct {
	ExaPID  int    `gorm:"column:exapid;primaryKey;autoIncrement" json:"exapid"`
	WordID  *uint  `gorm:"column:wordid" json:"wordid"`
	EN      string `gorm:"column:en" json:"en"`
	CN      string `gorm:"column:cn" json:"cn"`
	Heat    *int   `gorm:"column:heat" json:"heat"`
	AddDate string `gorm:"column:adddate" json:"adddate"`
}

func (VocExample) TableName() string { return "tb_voc_examples" }
