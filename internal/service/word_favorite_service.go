package service

import (
	"errors"
	"strings"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"

	"gorm.io/gorm"
)

var ErrWordNotFound = errors.New("word not found")

type WordFavoriteService struct {
	db *gorm.DB
}

func NewWordFavoriteService(db *gorm.DB) *WordFavoriteService {
	return &WordFavoriteService{db: db}
}

func (s *WordFavoriteService) Add(userID int64, req dto.WordFavoriteRequest) error {
	if s.db == nil {
		return ErrDatabaseDisabled
	}

	wordID, err := s.resolveWordID(req.WordID, req.Word)
	if err != nil {
		return err
	}

	favorite := model.UserWordFavorite{UserID: userID, WordID: wordID, PaperID: req.PaperID}
	return s.db.Where("user_id = ? and wordid = ?", userID, wordID).FirstOrCreate(&favorite).Error
}

func (s *WordFavoriteService) Remove(userID int64, req dto.WordFavoriteRequest) error {
	if s.db == nil {
		return ErrDatabaseDisabled
	}

	wordID, err := s.resolveWordID(req.WordID, req.Word)
	if errors.Is(err, ErrWordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.db.Where("user_id = ? and wordid = ?", userID, wordID).Delete(&model.UserWordFavorite{}).Error
}

func (s *WordFavoriteService) Has(userID int64, req dto.WordFavoriteRequest) (bool, error) {
	if s.db == nil {
		return false, ErrDatabaseDisabled
	}

	wordID, err := s.resolveWordID(req.WordID, req.Word)
	if errors.Is(err, ErrWordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var count int64
	err = s.db.Model(&model.UserWordFavorite{}).Where("user_id = ? and wordid = ?", userID, wordID).Count(&count).Error
	return count > 0, err
}

func (s *WordFavoriteService) List(userID int64, req dto.WordFavoriteListRequest) (dto.CollectedWordList, error) {
	if s.db == nil {
		return dto.CollectedWordList{}, ErrDatabaseDisabled
	}

	req.Normalize()
	query := s.db.Table("user_word_favorites as f").
		Joins("inner join tb_vocabulary v on v.wordid = f.wordid").
		Where("f.user_id = ?", userID)
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("v.spelling like ? or v.paraphrase like ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return dto.CollectedWordList{}, err
	}

	var list []dto.CollectedWordItem
	err := query.Select(`v.wordid as id, f.id as favorite_id, v.wordid, v.spelling as word, v.spelling,
		v.UKphonetic as uk_phonetic, v.USphonetic as us_phonetic, v.paraphrase, v.frequency,
		f.paper_id, f.created_at`).
		Order("f.created_at desc, f.id desc").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Scan(&list).Error

	return dto.CollectedWordList{Total: total, Data: list}, err
}

func (s *WordFavoriteService) resolveWordID(wordID uint, word string) (uint, error) {
	if wordID > 0 {
		var count int64
		if err := s.db.Model(&model.Vocabulary{}).Where("wordid = ?", wordID).Count(&count).Error; err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, ErrWordNotFound
		}
		return wordID, nil
	}

	spelling := strings.TrimSpace(word)
	if spelling == "" {
		return 0, ErrInvalidParam
	}

	var resolved uint
	err := s.db.Table("tb_vocabulary").Select("wordid").Where("spelling = ?", spelling).Limit(1).Scan(&resolved).Error
	if err != nil {
		return 0, err
	}
	if resolved == 0 {
		return 0, ErrWordNotFound
	}
	return resolved, nil
}
