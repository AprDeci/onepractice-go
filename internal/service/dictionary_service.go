package service

import (
	"errors"
	"strings"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"

	"gorm.io/gorm"
)

const (
	dictionaryWordListSelect = "v.wordid, v.spelling, v.UKphonetic as uk_phonetic, v.USphonetic as us_phonetic, v.paraphrase, v.frequency"
	dictionaryWordListOrder  = "v.frequency desc, v.wordid desc"
)

type DictionaryService struct {
	db *gorm.DB
}

func NewDictionaryService(db *gorm.DB) *DictionaryService {
	return &DictionaryService{db: db}
}

func (s *DictionaryService) ListWords(req dto.DictionaryWordListRequest) (dto.PageListResult[dto.DictionaryWordListItem], error) {
	if s.db == nil {
		return dto.PageListResult[dto.DictionaryWordListItem]{}, ErrDatabaseDisabled
	}

	req.Normalize()
	query := s.buildWordListQuery(req)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return dto.PageListResult[dto.DictionaryWordListItem]{}, err
	}

	var words []dto.DictionaryWordListItem
	err := query.Select(dictionaryWordListSelect).
		Order(dictionaryWordListOrder).
		Offset(req.Offset()).
		Limit(req.PageSize).
		Scan(&words).Error

	return dto.PageListResult[dto.DictionaryWordListItem]{
		List: words, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, err
}

func (s *DictionaryService) LookupMeanings(req dto.DictionaryLookupRequest) (dto.DictionaryLookupResult, error) {
	if s.db == nil {
		return dto.DictionaryLookupResult{}, ErrDatabaseDisabled
	}

	req.Normalize()
	spelling := strings.TrimSpace(req.Spelling)
	query := s.db.Table("tb_vocabulary as v").Select(dictionaryWordListSelect)
	if req.Exact {
		query = query.Where("v.spelling = ?", spelling)
	} else {
		query = query.Where("v.spelling like ?", "%"+spelling+"%")
	}

	var words []dto.DictionaryWordListItem
	err := query.Order("v.frequency desc").Limit(req.Limit).Scan(&words).Error
	return dto.DictionaryLookupResult{
		Spelling: req.Spelling,
		Exact:    req.Exact,
		Total:    len(words),
		Items:    words,
	}, err
}

func (s *DictionaryService) GetWordBySpelling(spelling string) (dto.DictionaryWordDetail, error) {
	if s.db == nil {
		return dto.DictionaryWordDetail{}, ErrDatabaseDisabled
	}

	var wordID uint
	err := s.db.Table("tb_vocabulary").
		Select("wordid").
		Where("spelling = ?", strings.TrimSpace(spelling)).
		Limit(1).
		Scan(&wordID).Error
	if err != nil {
		return dto.DictionaryWordDetail{}, err
	}
	if wordID == 0 {
		return dto.DictionaryWordDetail{}, gorm.ErrRecordNotFound
	}

	return s.GetWordDetail(wordID)
}

func (s *DictionaryService) GetWordDetail(wordID uint) (dto.DictionaryWordDetail, error) {
	if s.db == nil {
		return dto.DictionaryWordDetail{}, ErrDatabaseDisabled
	}

	var voc model.Vocabulary
	err := s.db.Where("wordid = ?", wordID).First(&voc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.DictionaryWordDetail{}, gorm.ErrRecordNotFound
	}
	if err != nil {
		return dto.DictionaryWordDetail{}, err
	}

	result := dto.DictionaryWordDetail{
		Word: dto.DictionaryWordListItem{
			WordID:     voc.WordID,
			Spelling:   voc.Spelling,
			UKPhonetic: voc.UKPhonetic,
			USPhonetic: voc.USPhonetic,
			Paraphrase: voc.Paraphrase,
			Frequency:  voc.Frequency,
		},
	}

	if err = s.db.Table("tb_book as b").
		Select("b.bookid, b.bookname").
		Joins("inner join tb_voc_book vb on vb.bookid = b.bookid").
		Where("vb.wordid = ?", wordID).
		Order("b.bookid asc").
		Scan(&result.Books).Error; err != nil {
		return dto.DictionaryWordDetail{}, err
	}

	if err = s.db.Table("tb_voc_examples").
		Select("exapid, en, cn, heat, adddate").
		Where("wordid = ?", wordID).
		Order("heat desc, exapid asc").
		Limit(20).
		Scan(&result.Examples).Error; err != nil {
		return dto.DictionaryWordDetail{}, err
	}

	return result, nil
}

func (s *DictionaryService) ListBooks(req dto.DictionaryBookListRequest) (dto.PageListResult[dto.DictionaryBookListItem], error) {
	if s.db == nil {
		return dto.PageListResult[dto.DictionaryBookListItem]{}, ErrDatabaseDisabled
	}

	req.Normalize()
	query := s.db.Table("tb_book as b")
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("b.bookname like ?", "%"+keyword+"%")
	}
	if req.Status != nil {
		query = query.Where("b.status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return dto.PageListResult[dto.DictionaryBookListItem]{}, err
	}

	var books []dto.DictionaryBookListItem
	err := query.Select("b.bookid, b.bookname, b.voccount, b.status").
		Order("b.bookid asc").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Scan(&books).Error

	return dto.PageListResult[dto.DictionaryBookListItem]{
		List: books, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, err
}

func (s *DictionaryService) ListBookWords(bookID uint, req dto.DictionaryBookWordsRequest) (dto.PageListResult[dto.DictionaryWordListItem], error) {
	if s.db == nil {
		return dto.PageListResult[dto.DictionaryWordListItem]{}, ErrDatabaseDisabled
	}

	req.Normalize()
	query := s.db.Table("tb_vocabulary as v").
		Where("EXISTS (SELECT 1 FROM tb_voc_book vb WHERE vb.wordid = v.wordid AND vb.bookid = ?)", bookID)
	query = applyDictionaryKeywordFilter(query, req.Keyword)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return dto.PageListResult[dto.DictionaryWordListItem]{}, err
	}

	var words []dto.DictionaryWordListItem
	err := query.Select(dictionaryWordListSelect).
		Order(dictionaryWordListOrder).
		Offset(req.Offset()).
		Limit(req.PageSize).
		Scan(&words).Error

	return dto.PageListResult[dto.DictionaryWordListItem]{
		List: words, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, err
}

func (s *DictionaryService) buildWordListQuery(req dto.DictionaryWordListRequest) *gorm.DB {
	query := s.db.Table("tb_vocabulary as v")
	if req.BookID != nil {
		query = query.Where("EXISTS (SELECT 1 FROM tb_voc_book vb WHERE vb.wordid = v.wordid AND vb.bookid = ?)", *req.BookID)
	}

	query = applyDictionaryKeywordFilter(query, req.Keyword)
	if spelling := strings.TrimSpace(req.Spelling); spelling != "" {
		query = query.Where("v.spelling like ?", "%"+spelling+"%")
	}
	if paraphrase := strings.TrimSpace(req.Paraphrase); paraphrase != "" {
		query = query.Where("v.paraphrase like ?", "%"+paraphrase+"%")
	}
	if req.MinFrequency != nil {
		query = query.Where("v.frequency >= ?", *req.MinFrequency)
	}
	if req.MaxFrequency != nil {
		query = query.Where("v.frequency <= ?", *req.MaxFrequency)
	}

	return query
}

func applyDictionaryKeywordFilter(query *gorm.DB, keyword string) *gorm.DB {
	if keyword := strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("v.spelling like ? or v.paraphrase like ?", like, like)
	}
	return query
}
