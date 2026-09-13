package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"onepractice-golang/internal/config"
	"onepractice-golang/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PointActionOCR        = "ocr"
	PointActionEssay      = "essay"
	PointActionDailyLogin = "daily_login"

	PointTypeOCRSpend    = "ocr_spend"
	PointTypeOCRRefund   = "ocr_refund"
	PointTypeEssaySpend  = "essay_spend"
	PointTypeEssayRefund = "essay_refund"
	PointTypeDailyLogin  = "daily_login"

	PointStatusDone     = 0
	PointStatusReserved = 1
	PointStatusSettled  = 2
	PointStatusRefunded = 3
)

// checkinKeyTTL 覆盖整个自然日，跨天后由新的 key 自然区分。
const checkinKeyTTL = 48 * time.Hour

// ErrInsufficientPoints 表示用户积分余额不足以完成扣费。
var ErrInsufficientPoints = errors.New("积分不足")

// errUnknownPointAction 表示 CostOf 收到了不受支持的动作。
var errUnknownPointAction = errors.New("unknown points action")

// pointLocation 固定使用东八区，避免依赖 Windows 上缺失的 tzdata。
var pointLocation = time.FixedZone("CST", 8*60*60)

// PointsService 负责积分余额、流水、规则与签到。redis 为可选增强，
// 为空或不可用时全部退化为数据库路径，绝不返回 ErrRedisDisabled。
type PointsService struct {
	db    *gorm.DB
	redis *redis.Client
	cfg   config.PointsConfig
}

func NewPointsService(db *gorm.DB, redisClient *redis.Client, cfg config.PointsConfig) *PointsService {
	return &PointsService{db: db, redis: redisClient, cfg: cfg}
}

// CostOf 返回某动作的积分单价：优先取生效中的 point_rules，否则回落配置默认值。
func (s *PointsService) CostOf(ctx context.Context, action string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDatabaseDisabled
	}

	// 时间区间比较必须与写入时区一致：GORM 与规则均按进程本地时区落库，
	// 若这里改用固定的 pointLocation，sqlite 以带 offset 的文本做字典序比较会错排。
	now := time.Now()
	var rule model.PointRule
	err := s.db.WithContext(ctx).
		Where("action = ? and enabled = ? and effective_from <= ? and (effective_to is null or effective_to >= ?)",
			action, true, now, now).
		Order("effective_from desc").
		First(&rule).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if err == nil && rule.Value > 0 {
		return rule.Value, nil
	}

	switch action {
	case PointActionOCR:
		return s.cfg.Defaults.OCRCost, nil
	case PointActionEssay:
		return s.cfg.Defaults.EssayCost, nil
	case PointActionDailyLogin:
		return s.cfg.Defaults.DailyLoginReward, nil
	default:
		return 0, fmt.Errorf("%w: %s", errUnknownPointAction, action)
	}
}

// Deduct 在单个事务内记一笔扣费流水并条件扣减余额。流水先落库，
// 余额不足时事务回滚从而移除流水，保证不产生悬空记录。重复 bizID 幂等返回 nil。
func (s *PointsService) Deduct(ctx context.Context, userID, amount int64, typ, bizID, remark string) error {
	if s == nil || s.db == nil {
		return ErrDatabaseDisabled
	}

	status := PointStatusDone
	if typ == PointTypeEssaySpend {
		status = PointStatusReserved
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txn := model.PointTransaction{
			UserID: userID,
			Delta:  -amount,
			Type:   typ,
			BizID:  bizID,
			Status: status,
			Remark: remark,
		}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&txn)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}

		upd := tx.Model(&model.UserPoints{}).
			Where("user_id = ? and balance >= ?", userID, amount).
			UpdateColumn("balance", gorm.Expr("balance - ?", amount))
		if upd.Error != nil {
			return upd.Error
		}
		if upd.RowsAffected == 0 {
			return ErrInsufficientPoints
		}

		balance, err := s.readBalance(tx, userID)
		if err != nil {
			return err
		}
		return tx.Model(&model.PointTransaction{}).
			Where("id = ?", txn.ID).
			UpdateColumn("balance_after", balance).Error
	})
}

// Grant 在单个事务内记一笔入账流水并累加余额（不存在则懒创建）。
// 重复 bizID 幂等返回 (false, nil)。
func (s *PointsService) Grant(ctx context.Context, userID, amount int64, typ, bizID, remark string) (bool, error) {
	if s == nil || s.db == nil {
		return false, ErrDatabaseDisabled
	}

	granted := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txn := model.PointTransaction{
			UserID: userID,
			Delta:  amount,
			Type:   typ,
			BizID:  bizID,
			Status: PointStatusDone,
			Remark: remark,
		}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&txn)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}

		if err := s.addBalance(tx, userID, amount); err != nil {
			return err
		}
		balance, err := s.readBalance(tx, userID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.PointTransaction{}).
			Where("id = ?", txn.ID).
			UpdateColumn("balance_after", balance).Error; err != nil {
			return err
		}
		granted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return granted, nil
}

// Refund 依据原始扣费流水发起退款：金额取反，追加退款流水并把原流水标记为已退款。
// 原始流水不存在或已退款时幂等返回 nil。
func (s *PointsService) Refund(ctx context.Context, userID int64, spendType, bizID, remark string) error {
	if s == nil || s.db == nil {
		return ErrDatabaseDisabled
	}

	refundType, ok := refundTypeOf(spendType)
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownPointAction, spendType)
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var spend model.PointTransaction
		err := tx.Where("user_id = ? and type = ? and biz_id = ?", userID, spendType, bizID).
			First(&spend).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		amount := -spend.Delta
		if amount <= 0 {
			return nil
		}

		refund := model.PointTransaction{
			UserID: userID,
			Delta:  amount,
			Type:   refundType,
			BizID:  bizID,
			Status: PointStatusDone,
			Remark: remark,
		}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&refund)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}

		if err := s.addBalance(tx, userID, amount); err != nil {
			return err
		}
		balance, err := s.readBalance(tx, userID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.PointTransaction{}).
			Where("id = ?", refund.ID).
			UpdateColumn("balance_after", balance).Error; err != nil {
			return err
		}
		return tx.Model(&model.PointTransaction{}).
			Where("id = ?", spend.ID).
			UpdateColumn("status", PointStatusRefunded).Error
	})
}

// Settle 把预留中的作文扣费结算为已消费。
func (s *PointsService) Settle(ctx context.Context, userID int64, bizID string) error {
	if s == nil || s.db == nil {
		return ErrDatabaseDisabled
	}
	return s.db.WithContext(ctx).Model(&model.PointTransaction{}).
		Where("user_id = ? and type = ? and biz_id = ? and status = ?",
			userID, PointTypeEssaySpend, bizID, PointStatusReserved).
		UpdateColumn("status", PointStatusSettled).Error
}

// Balance 返回用户余额，无记录时返回 0。
func (s *PointsService) Balance(ctx context.Context, userID int64) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDatabaseDisabled
	}
	var row model.UserPoints
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return row.Balance, nil
}

// ListTransactions 分页返回用户的积分流水（按时间倒序）。
func (s *PointsService) ListTransactions(ctx context.Context, userID int64, offset, limit int) ([]model.PointTransaction, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, ErrDatabaseDisabled
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&model.PointTransaction{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.PointTransaction
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc, id desc").
		Offset(offset).Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// CheckinStatus 判断用户当日是否已签到：先用 Redis，未命中再查库并回填缓存。
// 负数结果不缓存，避免漏签被错误抑制。
func (s *PointsService) CheckinStatus(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.db == nil {
		return false, ErrDatabaseDisabled
	}

	now := time.Now().In(pointLocation)
	key := checkinKey(userID, now)
	if s.redis != nil {
		if n, err := s.redis.Exists(ctx, key).Result(); err == nil && n > 0 {
			return true, nil
		}
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&model.PointTransaction{}).
		Where("user_id = ? and type = ? and biz_id = ?", userID, PointTypeDailyLogin, loginBizID(now)).
		Count(&count).Error; err != nil {
		return false, err
	}
	checkedIn := count > 0
	if checkedIn && s.redis != nil {
		_ = s.redis.Set(ctx, key, "1", checkinKeyTTL).Err()
	}
	return checkedIn, nil
}

// ClaimDailyLogin 领取每日登录奖励，同日重复领取幂等返回 (false, 当前余额)。
func (s *PointsService) ClaimDailyLogin(ctx context.Context, userID int64) (granted bool, balance int64, err error) {
	if s == nil || s.db == nil {
		return false, 0, ErrDatabaseDisabled
	}

	cost, err := s.CostOf(ctx, PointActionDailyLogin)
	if err != nil {
		return false, 0, err
	}
	now := time.Now().In(pointLocation)
	granted, err = s.Grant(ctx, userID, cost, PointTypeDailyLogin, loginBizID(now), "每日签到")
	if err != nil {
		return false, 0, err
	}
	if granted && s.redis != nil {
		_ = s.redis.Set(ctx, checkinKey(userID, now), "1", checkinKeyTTL).Err()
	}

	balance, err = s.Balance(ctx, userID)
	if err != nil {
		return granted, 0, err
	}
	return granted, balance, nil
}

// SweepExpiredReserved 扫描超时未结算的预扣流水并逐一退款，返回退款笔数。
func (s *PointsService) SweepExpiredReserved(ctx context.Context) (int, error) {
	if s == nil || s.db == nil {
		return 0, ErrDatabaseDisabled
	}

	timeout := s.cfg.ReserveTimeoutMins
	if timeout <= 0 {
		timeout = 10
	}
	// 与 created_at 的写入时区保持一致，理由同 CostOf。
	cutoff := time.Now().Add(-time.Duration(timeout) * time.Minute)

	var rows []model.PointTransaction
	if err := s.db.WithContext(ctx).
		Where("type = ? and status = ? and created_at < ?",
			PointTypeEssaySpend, PointStatusReserved, cutoff).
		Find(&rows).Error; err != nil {
		return 0, err
	}

	refunded := 0
	for _, row := range rows {
		if err := s.Refund(ctx, row.UserID, PointTypeEssaySpend, row.BizID, "超时未结算自动退款"); err != nil {
			return refunded, err
		}
		refunded++
	}
	return refunded, nil
}

// SpendStatus 查询某笔消费流水的状态；未找到时返回 found=false。
func (s *PointsService) SpendStatus(ctx context.Context, userID int64, typ, bizID string) (int, bool, error) {
	if s == nil || s.db == nil {
		return 0, false, ErrDatabaseDisabled
	}
	var row model.PointTransaction
	err := s.db.WithContext(ctx).
		Where("user_id = ? and type = ? and biz_id = ?", userID, typ, bizID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return row.Status, true, nil
}

// addBalance 原子累加余额，余额行不存在时懒创建。
func (s *PointsService) addBalance(tx *gorm.DB, userID, amount int64) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"balance":    gorm.Expr("balance + ?", amount),
			"updated_at": time.Now(),
		}),
	}).Create(&model.UserPoints{UserID: userID, Balance: amount}).Error
}

// readBalance 在事务内读取用户当前余额。
func (s *PointsService) readBalance(tx *gorm.DB, userID int64) (int64, error) {
	var balance int64
	err := tx.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		Select("balance").
		Scan(&balance).Error
	return balance, err
}

func refundTypeOf(spendType string) (string, bool) {
	switch spendType {
	case PointTypeOCRSpend:
		return PointTypeOCRRefund, true
	case PointTypeEssaySpend:
		return PointTypeEssayRefund, true
	default:
		return "", false
	}
}

func loginBizID(t time.Time) string {
	return "login:" + t.In(pointLocation).Format("2006-01-02")
}

func checkinKey(userID int64, t time.Time) string {
	return fmt.Sprintf("onepractice:points:checkin:%d:%s", userID, t.In(pointLocation).Format("20060102"))
}
