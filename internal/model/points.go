package model

import "time"

// UserPoints 保存用户当前积分余额，user_id 既是主键也是用户标识。
type UserPoints struct {
	UserID    int64     `gorm:"column:user_id;primaryKey" json:"userId"`
	Balance   int64     `gorm:"column:balance" json:"balance"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (UserPoints) TableName() string { return "user_points" }

// PointTransaction 是积分流水的追加记录，(user_id,type,biz_id) 唯一用于幂等去重。
type PointTransaction struct {
	ID           uint64    `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id;uniqueIndex:uk_pt_idem" json:"userId"`
	Delta        int64     `gorm:"column:delta" json:"delta"`
	BalanceAfter int64     `gorm:"column:balance_after" json:"balanceAfter"`
	Type         string    `gorm:"column:type;uniqueIndex:uk_pt_idem" json:"type"`
	BizID        string    `gorm:"column:biz_id;uniqueIndex:uk_pt_idem" json:"bizId"`
	Status       int       `gorm:"column:status" json:"status"`
	Remark       string    `gorm:"column:remark" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (PointTransaction) TableName() string { return "point_transactions" }

// PointRule 描述某个动作的积分单价及生效时间区间，未命中时回落到配置默认值。
type PointRule struct {
	ID            uint64     `gorm:"column:id;primaryKey" json:"id"`
	Action        string     `gorm:"column:action" json:"action"`
	Value         int64      `gorm:"column:value" json:"value"`
	EffectiveFrom time.Time  `gorm:"column:effective_from" json:"effectiveFrom"`
	EffectiveTo   *time.Time `gorm:"column:effective_to" json:"effectiveTo"`
	Enabled       bool       `gorm:"column:enabled" json:"enabled"`
	Remark        string     `gorm:"column:remark" json:"remark"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (PointRule) TableName() string { return "point_rules" }
