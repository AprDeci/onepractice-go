package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type RecordRequest struct {
	RecordID     string      `json:"recordId"`
	PaperID      StringInt   `json:"paperId" binding:"required"`
	Type         string      `json:"type"`
	IsFinished   int         `json:"isfinished"`
	Answers      string      `json:"answers"`
	Score        int         `json:"score"`
	TotalScore   int         `json:"totalscore"`
	TimeSpend    int         `json:"timespend"`
	HasSpendTime StringInt64 `json:"hasspendtime"`
}

type StringInt int

func (v *StringInt) UnmarshalJSON(data []byte) error {
	parsed, err := parseJSONInt(data, 0)
	if err != nil {
		return err
	}
	*v = StringInt(parsed)
	return nil
}

type StringInt64 int64

func (v *StringInt64) UnmarshalJSON(data []byte) error {
	parsed, err := parseJSONInt(data, 64)
	if err != nil {
		return err
	}
	*v = StringInt64(parsed)
	return nil
}

func parseJSONInt(data []byte, bitSize int) (int64, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, err
	}
	switch value := raw.(type) {
	case nil:
		return 0, nil
	case float64:
		return int64(value), nil
	case string:
		if value == "" {
			return 0, nil
		}
		return strconv.ParseInt(value, 10, bitSize)
	default:
		return 0, fmt.Errorf("unsupported integer value type %T", raw)
	}
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
