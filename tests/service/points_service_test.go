package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"onepractice-golang/internal/config"
	"onepractice-golang/internal/model"
	"onepractice-golang/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func testPointsConfig() config.PointsConfig {
	return config.PointsConfig{
		Defaults: config.PointsDefaultsConfig{
			OCRCost:          1,
			EssayCost:        2,
			DailyLoginReward: 6,
		},
		ReserveTimeoutMins: 10,
	}
}

// newPointsTestDB 为每个用例创建独立的内存 sqlite 库，并串行化连接以避免
// shared-cache 下的写锁竞争。
func newPointsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:pt_%s?mode=memory&cache=shared", name)), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.UserPoints{}, &model.PointTransaction{}, &model.PointRule{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func countTransactions(t *testing.T, db *gorm.DB, where string, args ...any) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&model.PointTransaction{}).Where(where, args...).Count(&count).Error; err != nil {
		t.Fatalf("count transactions: %v", err)
	}
	return count
}

func TestPointsServiceNilDatabase(t *testing.T) {
	svc := service.NewPointsService(nil, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.CostOf(ctx, service.PointActionOCR); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("CostOf() error = %v, want ErrDatabaseDisabled", err)
	}
	if err := svc.Deduct(ctx, 1, 1, service.PointTypeOCRSpend, "b", ""); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Deduct() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, err := svc.Grant(ctx, 1, 1, service.PointTypeDailyLogin, "b", ""); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Grant() error = %v, want ErrDatabaseDisabled", err)
	}
	if err := svc.Refund(ctx, 1, service.PointTypeOCRSpend, "b", ""); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Refund() error = %v, want ErrDatabaseDisabled", err)
	}
	if err := svc.Settle(ctx, 1, "b"); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Settle() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, err := svc.Balance(ctx, 1); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Balance() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, _, err := svc.ListTransactions(ctx, 1, 0, 10); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("ListTransactions() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, err := svc.CheckinStatus(ctx, 1); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("CheckinStatus() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, _, err := svc.ClaimDailyLogin(ctx, 1); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("ClaimDailyLogin() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, err := svc.SweepExpiredReserved(ctx); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("SweepExpiredReserved() error = %v, want ErrDatabaseDisabled", err)
	}
}

func TestPointsConfigValidate(t *testing.T) {
	if err := testPointsConfig().Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	bad := testPointsConfig()
	bad.Defaults.OCRCost = 0
	if err := bad.Validate(); err == nil || !strings.Contains(err.Error(), "points.defaults") {
		t.Fatalf("Validate() error = %v, want points.defaults mention", err)
	}
}

func TestPointsCostOfFallsBackToDefaults(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	cases := map[string]int64{
		service.PointActionOCR:        1,
		service.PointActionEssay:      2,
		service.PointActionDailyLogin: 6,
	}
	for action, want := range cases {
		got, err := svc.CostOf(ctx, action)
		if err != nil {
			t.Fatalf("CostOf(%q) error = %v", action, err)
		}
		if got != want {
			t.Fatalf("CostOf(%q) = %d, want %d", action, got, want)
		}
	}

	if _, err := svc.CostOf(ctx, "unknown"); err == nil {
		t.Fatal("CostOf(unknown) error = nil, want error")
	}
}

func TestPointsCostOfActiveRuleOverrides(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	now := time.Now()
	mustCreateRule(t, db, model.PointRule{Action: service.PointActionOCR, Value: 5, EffectiveFrom: now.Add(-2 * time.Hour), Enabled: true})
	// 更晚生效的规则应优先。
	mustCreateRule(t, db, model.PointRule{Action: service.PointActionOCR, Value: 9, EffectiveFrom: now.Add(-time.Hour), Enabled: true})

	got, err := svc.CostOf(ctx, service.PointActionOCR)
	if err != nil {
		t.Fatalf("CostOf() error = %v", err)
	}
	if got != 9 {
		t.Fatalf("CostOf() = %d, want newest active rule 9", got)
	}
}

func TestPointsCostOfIgnoresDisabledExpiredFutureRules(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	now := time.Now()
	expiredTo := now.Add(-time.Minute)
	mustCreateRule(t, db, model.PointRule{Action: service.PointActionOCR, Value: 11, EffectiveFrom: now.Add(-2 * time.Hour), Enabled: false})
	mustCreateRule(t, db, model.PointRule{Action: service.PointActionOCR, Value: 12, EffectiveFrom: now.Add(-2 * time.Hour), EffectiveTo: &expiredTo, Enabled: true})
	mustCreateRule(t, db, model.PointRule{Action: service.PointActionOCR, Value: 13, EffectiveFrom: now.Add(time.Hour), Enabled: true})

	got, err := svc.CostOf(ctx, service.PointActionOCR)
	if err != nil {
		t.Fatalf("CostOf() error = %v", err)
	}
	if got != 1 {
		t.Fatalf("CostOf() = %d, want fallback default 1", got)
	}
}

func TestPointsDeductAndIdempotentReplay(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.Grant(ctx, 1, 10, service.PointTypeDailyLogin, "seed", ""); err != nil {
		t.Fatalf("Grant(seed) error = %v", err)
	}

	if err := svc.Deduct(ctx, 1, 3, service.PointTypeOCRSpend, "ocr-1", "识别"); err != nil {
		t.Fatalf("Deduct() error = %v", err)
	}
	assertBalance(t, svc, 1, 7)
	if count := countTransactions(t, db, "user_id = ? and biz_id = ?", 1, "ocr-1"); count != 1 {
		t.Fatalf("ledger count = %d, want 1", count)
	}
	var row model.PointTransaction
	if err := db.Where("user_id = ? and biz_id = ?", 1, "ocr-1").First(&row).Error; err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if row.Delta != -3 || row.BalanceAfter != 7 || row.Status != service.PointStatusDone {
		t.Fatalf("ledger = %+v, want delta=-3 balanceAfter=7 status=done", row)
	}

	// 幂等重放：不重复扣费，也不新增流水。
	if err := svc.Deduct(ctx, 1, 3, service.PointTypeOCRSpend, "ocr-1", "识别"); err != nil {
		t.Fatalf("Deduct replay error = %v", err)
	}
	assertBalance(t, svc, 1, 7)
	if count := countTransactions(t, db, "user_id = ? and biz_id = ?", 1, "ocr-1"); count != 1 {
		t.Fatalf("ledger count after replay = %d, want 1", count)
	}

	// 余额不足：返回哨兵错误，事务回滚不留下悬空流水。
	if err := svc.Deduct(ctx, 1, 100, service.PointTypeOCRSpend, "ocr-broke", ""); !errors.Is(err, service.ErrInsufficientPoints) {
		t.Fatalf("Deduct(insufficient) error = %v, want ErrInsufficientPoints", err)
	}
	assertBalance(t, svc, 1, 7)
	if count := countTransactions(t, db, "user_id = ? and biz_id = ?", 1, "ocr-broke"); count != 0 {
		t.Fatalf("dangling ledger count = %d, want 0", count)
	}
}

func TestPointsDeductConcurrentOnlyOneSucceeds(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.Grant(ctx, 1, 10, service.PointTypeDailyLogin, "seed", ""); err != nil {
		t.Fatalf("Grant(seed) error = %v", err)
	}

	bizIDs := []string{"race-a", "race-b"}
	errs := make([]error, len(bizIDs))
	var wg sync.WaitGroup
	for i, bizID := range bizIDs {
		wg.Add(1)
		go func(i int, bizID string) {
			defer wg.Done()
			errs[i] = svc.Deduct(ctx, 1, 10, service.PointTypeOCRSpend, bizID, "")
		}(i, bizID)
	}
	wg.Wait()

	success, insufficient := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, service.ErrInsufficientPoints):
			insufficient++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || insufficient != 1 {
		t.Fatalf("success=%d insufficient=%d, want 1/1", success, insufficient)
	}
	assertBalance(t, svc, 1, 0)
}

func TestPointsGrantLazyCreateAndIdempotent(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	granted, err := svc.Grant(ctx, 1, 5, service.PointTypeDailyLogin, "g1", "奖励")
	if err != nil {
		t.Fatalf("Grant() error = %v", err)
	}
	if !granted {
		t.Fatal("Grant() granted = false, want true")
	}
	assertBalance(t, svc, 1, 5)

	granted, err = svc.Grant(ctx, 1, 5, service.PointTypeDailyLogin, "g1", "奖励")
	if err != nil {
		t.Fatalf("Grant replay error = %v", err)
	}
	if granted {
		t.Fatal("Grant replay granted = true, want false")
	}
	assertBalance(t, svc, 1, 5)
	if count := countTransactions(t, db, "user_id = ?", 1); count != 1 {
		t.Fatalf("ledger count = %d, want 1", count)
	}
}

func TestPointsRefundIdempotentAndRestoresBalance(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.Grant(ctx, 1, 10, service.PointTypeDailyLogin, "seed", ""); err != nil {
		t.Fatalf("Grant(seed) error = %v", err)
	}
	if err := svc.Deduct(ctx, 1, 4, service.PointTypeEssaySpend, "essay-1", ""); err != nil {
		t.Fatalf("Deduct() error = %v", err)
	}
	assertBalance(t, svc, 1, 6)

	if err := svc.Refund(ctx, 1, service.PointTypeEssaySpend, "essay-1", "退款"); err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	assertBalance(t, svc, 1, 10)

	var spend model.PointTransaction
	if err := db.Where("user_id = ? and type = ? and biz_id = ?", 1, service.PointTypeEssaySpend, "essay-1").First(&spend).Error; err != nil {
		t.Fatalf("load spend: %v", err)
	}
	if spend.Status != service.PointStatusRefunded {
		t.Fatalf("spend status = %d, want refunded", spend.Status)
	}
	var refund model.PointTransaction
	if err := db.Where("user_id = ? and type = ? and biz_id = ?", 1, service.PointTypeEssayRefund, "essay-1").First(&refund).Error; err != nil {
		t.Fatalf("load refund: %v", err)
	}
	if refund.Delta != 4 {
		t.Fatalf("refund delta = %d, want 4", refund.Delta)
	}

	// 幂等重放不重复退款。
	if err := svc.Refund(ctx, 1, service.PointTypeEssaySpend, "essay-1", "退款"); err != nil {
		t.Fatalf("Refund replay error = %v", err)
	}
	assertBalance(t, svc, 1, 10)
	if count := countTransactions(t, db, "user_id = ?", 1); count != 3 {
		t.Fatalf("ledger count = %d, want 3", count)
	}

	// 原流水不存在时幂等返回 nil。
	if err := svc.Refund(ctx, 1, service.PointTypeOCRSpend, "missing", ""); err != nil {
		t.Fatalf("Refund(missing) error = %v", err)
	}
}

func TestPointsSettleFlipsReserved(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.Grant(ctx, 1, 10, service.PointTypeDailyLogin, "seed", ""); err != nil {
		t.Fatalf("Grant(seed) error = %v", err)
	}
	if err := svc.Deduct(ctx, 1, 4, service.PointTypeEssaySpend, "essay-1", ""); err != nil {
		t.Fatalf("Deduct() error = %v", err)
	}

	var before model.PointTransaction
	if err := db.Where("biz_id = ?", "essay-1").First(&before).Error; err != nil {
		t.Fatalf("load before: %v", err)
	}
	if before.Status != service.PointStatusReserved {
		t.Fatalf("status before = %d, want reserved", before.Status)
	}

	if err := svc.Settle(ctx, 1, "essay-1"); err != nil {
		t.Fatalf("Settle() error = %v", err)
	}
	var after model.PointTransaction
	if err := db.Where("biz_id = ?", "essay-1").First(&after).Error; err != nil {
		t.Fatalf("load after: %v", err)
	}
	if after.Status != service.PointStatusSettled {
		t.Fatalf("status after = %d, want settled", after.Status)
	}
}

func TestPointsSweepExpiredReserved(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	if _, err := svc.Grant(ctx, 1, 20, service.PointTypeDailyLogin, "seed", ""); err != nil {
		t.Fatalf("Grant(seed) error = %v", err)
	}
	if err := svc.Deduct(ctx, 1, 4, service.PointTypeEssaySpend, "stale", ""); err != nil {
		t.Fatalf("Deduct(stale) error = %v", err)
	}
	// 把预扣流水回拨到超时之前。
	if err := db.Model(&model.PointTransaction{}).Where("biz_id = ?", "stale").
		UpdateColumn("created_at", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatalf("backdate: %v", err)
	}
	// 新预扣不应被扫到。
	if err := svc.Deduct(ctx, 1, 4, service.PointTypeEssaySpend, "fresh", ""); err != nil {
		t.Fatalf("Deduct(fresh) error = %v", err)
	}
	assertBalance(t, svc, 1, 12)

	refunded, err := svc.SweepExpiredReserved(ctx)
	if err != nil {
		t.Fatalf("SweepExpiredReserved() error = %v", err)
	}
	if refunded != 1 {
		t.Fatalf("refunded = %d, want 1", refunded)
	}
	assertBalance(t, svc, 1, 16)

	var stale model.PointTransaction
	if err := db.Where("biz_id = ?", "stale").First(&stale).Error; err != nil {
		t.Fatalf("load stale: %v", err)
	}
	if stale.Status != service.PointStatusRefunded {
		t.Fatalf("stale status = %d, want refunded", stale.Status)
	}

	// 再次扫描不应重复退款。
	refunded, err = svc.SweepExpiredReserved(ctx)
	if err != nil {
		t.Fatalf("SweepExpiredReserved() second error = %v", err)
	}
	if refunded != 0 {
		t.Fatalf("second refunded = %d, want 0", refunded)
	}
}

func TestPointsClaimDailyLoginIdempotent(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	granted, balance, err := svc.ClaimDailyLogin(ctx, 1)
	if err != nil {
		t.Fatalf("ClaimDailyLogin() error = %v", err)
	}
	if !granted || balance != 6 {
		t.Fatalf("ClaimDailyLogin() = (%v, %d), want (true, 6)", granted, balance)
	}

	granted, balance, err = svc.ClaimDailyLogin(ctx, 1)
	if err != nil {
		t.Fatalf("ClaimDailyLogin() second error = %v", err)
	}
	if granted || balance != 6 {
		t.Fatalf("ClaimDailyLogin() second = (%v, %d), want (false, 6)", granted, balance)
	}

	checkedIn, err := svc.CheckinStatus(ctx, 1)
	if err != nil {
		t.Fatalf("CheckinStatus() error = %v", err)
	}
	if !checkedIn {
		t.Fatal("CheckinStatus() = false, want true after claim (DB fallback)")
	}

	if count := countTransactions(t, db, "user_id = ? and type = ?", 1, service.PointTypeDailyLogin); count != 1 {
		t.Fatalf("daily_login ledger count = %d, want 1", count)
	}
}

func TestPointsCheckinStatusFalseBeforeClaim(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	checkedIn, err := svc.CheckinStatus(ctx, 1)
	if err != nil {
		t.Fatalf("CheckinStatus() error = %v", err)
	}
	if checkedIn {
		t.Fatal("CheckinStatus() = true, want false before claim")
	}
}

func TestPointsServiceIgnoresRedisErrors(t *testing.T) {
	db := newPointsTestDB(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 50 * time.Millisecond,
		MaxRetries:  -1,
	})
	defer redisClient.Close()
	svc := service.NewPointsService(db, redisClient, testPointsConfig())
	ctx := context.Background()

	// Redis 不可用不应阻断任何基础能力。
	granted, err := svc.Grant(ctx, 1, 10, service.PointTypeDailyLogin, "seed", "")
	if err != nil || !granted {
		t.Fatalf("Grant() = (%v, %v), want (true, nil)", granted, err)
	}
	if err := svc.Deduct(ctx, 1, 2, service.PointTypeOCRSpend, "ocr-1", ""); err != nil {
		t.Fatalf("Deduct() error = %v", err)
	}
	assertBalance(t, svc, 1, 8)

	checkedIn, err := svc.CheckinStatus(ctx, 1)
	if err != nil {
		t.Fatalf("CheckinStatus() error = %v", err)
	}
	if checkedIn {
		t.Fatal("CheckinStatus() = true, want false")
	}

	claimGranted, balance, err := svc.ClaimDailyLogin(ctx, 1)
	if err != nil {
		t.Fatalf("ClaimDailyLogin() error = %v", err)
	}
	if !claimGranted || balance != 14 {
		t.Fatalf("ClaimDailyLogin() = (%v, %d), want (true, 14)", claimGranted, balance)
	}

	// Redis SET 失败被忽略，数据库回退仍能识别已签到。
	checkedIn, err = svc.CheckinStatus(ctx, 1)
	if err != nil {
		t.Fatalf("CheckinStatus() after claim error = %v", err)
	}
	if !checkedIn {
		t.Fatal("CheckinStatus() after claim = false, want true")
	}
}

func TestPointsListTransactions(t *testing.T) {
	db := newPointsTestDB(t)
	svc := service.NewPointsService(db, nil, testPointsConfig())
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := svc.Grant(ctx, 1, int64(i+1), service.PointTypeDailyLogin, fmt.Sprintf("g%d", i), ""); err != nil {
			t.Fatalf("Grant(%d) error = %v", i, err)
		}
	}

	rows, total, err := svc.ListTransactions(ctx, 1, 0, 2)
	if err != nil {
		t.Fatalf("ListTransactions() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}

	// 无流水的用户返回空列表。
	rows, total, err = svc.ListTransactions(ctx, 2, 0, 10)
	if err != nil {
		t.Fatalf("ListTransactions(empty) error = %v", err)
	}
	if total != 0 || len(rows) != 0 {
		t.Fatalf("empty list = (%d, %d), want (0, 0)", total, len(rows))
	}
}

func mustCreateRule(t *testing.T, db *gorm.DB, rule model.PointRule) {
	t.Helper()
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("create rule: %v", err)
	}
}

func assertBalance(t *testing.T, svc *service.PointsService, userID, want int64) {
	t.Helper()
	got, err := svc.Balance(context.Background(), userID)
	if err != nil {
		t.Fatalf("Balance(%d) error = %v", userID, err)
	}
	if got != want {
		t.Fatalf("Balance(%d) = %d, want %d", userID, got, want)
	}
}
