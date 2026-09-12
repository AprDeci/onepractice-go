package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

// PointsHandler 处理积分余额、流水与签到接口。
type PointsHandler struct {
	service *service.PointsService
}

func NewPointsHandler(s *service.PointsService) *PointsHandler {
	return &PointsHandler{service: s}
}

func pointsError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInsufficientPoints):
		response.Error(c, apperror.New(apperror.CodeInsufficientPoints, "积分不足"))
	case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}

// Balance 查询当前用户积分余额。
// @Summary 查询积分余额
// @Description 返回当前登录用户的积分余额。
// @Tags points
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body{data=apiv1.PointsBalanceResponse}
// @Router /api/v1/points/balance [get]
func (h *PointsHandler) Balance(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	balance, err := h.service.Balance(c.Request.Context(), userID)
	if err != nil {
		pointsError(c, err)
		return
	}
	response.Success(c, dtoV1.PointsBalanceResponse{Balance: balance})
}

// ListTransactions 分页查询当前用户积分流水。
// @Summary 查询积分流水
// @Description 分页返回当前登录用户的积分流水，按时间倒序。
// @Tags points
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body{data=apiv1.PointsTransactionPage}
// @Router /api/v1/points/transactions [get]
func (h *PointsHandler) ListTransactions(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	q := pageQuery(c)
	rows, total, err := h.service.ListTransactions(c.Request.Context(), userID, q.Offset(), q.PageSize)
	if err != nil {
		pointsError(c, err)
		return
	}
	items := make([]dtoV1.PointsTransactionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dtoV1.PointsTransactionItem{
			ID:           row.ID,
			Delta:        row.Delta,
			BalanceAfter: row.BalanceAfter,
			Type:         row.Type,
			BizID:        row.BizID,
			Remark:       row.Remark,
			CreatedAt:    row.CreatedAt,
		})
	}
	response.Success(c, newPage(items, total, q))
}

// DailyCheckin 每日签到领取积分。
// @Summary 每日签到
// @Description 领取当日签到积分，同日重复领取幂等返回 granted=false。
// @Tags points
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body{data=apiv1.PointsGrantResponse}
// @Router /api/v1/points/daily-checkin [post]
func (h *PointsHandler) DailyCheckin(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	granted, balance, err := h.service.ClaimDailyLogin(c.Request.Context(), userID)
	if err != nil {
		pointsError(c, err)
		return
	}
	response.Success(c, dtoV1.PointsGrantResponse{Granted: granted, Balance: balance})
}

// CheckinStatus 查询今日是否已签到。
// @Summary 查询签到状态
// @Description 返回当前登录用户今日是否已签到。
// @Tags points
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body{data=apiv1.CheckinStatusResponse}
// @Router /api/v1/points/checkin-status [get]
func (h *PointsHandler) CheckinStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	checkedIn, err := h.service.CheckinStatus(c.Request.Context(), userID)
	if err != nil {
		pointsError(c, err)
		return
	}
	response.Success(c, dtoV1.CheckinStatusResponse{CheckedIn: checkedIn})
}

// Costs 返回各功能的当前积分消耗。
// @Summary 查询功能积分消耗
// @Description 返回 OCR、作文批改等功能的当前积分消耗，以及每日签到奖励。数值来自 point_rules 生效规则，缺失时回落配置默认值。
// @Tags points
// @Produce json
// @Success 200 {object} response.Body{data=apiv1.PointsCostsResponse}
// @Router /api/v1/points/costs [get]
func (h *PointsHandler) Costs(c *gin.Context) {
	ctx := c.Request.Context()

	ocrCost, err := h.service.CostOf(ctx, service.PointActionOCR)
	if err != nil {
		pointsError(c, err)
		return
	}
	essayCost, err := h.service.CostOf(ctx, service.PointActionEssay)
	if err != nil {
		pointsError(c, err)
		return
	}
	dailyReward, err := h.service.CostOf(ctx, service.PointActionDailyLogin)
	if err != nil {
		pointsError(c, err)
		return
	}

	response.Success(c, dtoV1.PointsCostsResponse{
		Costs: []dtoV1.PointsCostItem{
			{Action: service.PointActionOCR, Name: "OCR识别", Cost: ocrCost},
			{Action: service.PointActionEssay, Name: "作文批改", Cost: essayCost},
		},
		DailyLoginReward: dailyReward,
	})
}
