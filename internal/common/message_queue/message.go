package message_queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id          string      `json:"id"`
	CreateTime  time.Time   `json:"create_time"`
	ConsumeTime time.Time   `json:"consume_time"`
	Body        interface{} `json:"body"`
}

func NewMessage(id string, consumeTIme time.Time, body interface{}) *Message {
	if id == "" {
		if uid, err := uuid.NewV7(); err == nil {
			id = uid.String()
		}
	}
	return &Message{
		Id:          id,
		CreateTime:  time.Now(),
		ConsumeTime: consumeTIme,
		Body:        body,
	}
}

// GetScore 用于返回消息的分数
func (m *Message) GetScore() float64 {
	return float64(m.ConsumeTime.Unix())
}

// GetId 用于返回消息的ID
func (m *Message) GetId() string {
	return m.Id
}

// MarshalBinary 用于将消息结构体序列化为二进制数据
func (m *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary 用于将二进制数据反序列化为消息结构体
func (m *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}
