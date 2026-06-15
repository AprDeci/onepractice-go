package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Answer struct {
	Index  string `json:"index"`
	Answer string `json:"answer"`
}

type Option struct {
	Label   string `json:"label"`
	Content string `json:"content"`
}

type ReadItem struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type Answers []Answer
type Options []Option
type ReadItems []ReadItem
type StringList []string

func (v *Answers) Scan(value any) error        { return scanJSON(value, v) }
func (v Answers) Value() (driver.Value, error) { return valueJSON(v) }

func (v *Options) Scan(value any) error        { return scanJSON(value, v) }
func (v Options) Value() (driver.Value, error) { return valueJSON(v) }

func (v *ReadItems) Scan(value any) error        { return scanJSON(value, v) }
func (v ReadItems) Value() (driver.Value, error) { return valueJSON(v) }

func (v *StringList) Scan(value any) error        { return scanJSON(value, v) }
func (v StringList) Value() (driver.Value, error) { return valueJSON(v) }

func scanJSON(value any, target any) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported JSON value type %T", value)
	}

	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, target)
}

func valueJSON(value any) (driver.Value, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
