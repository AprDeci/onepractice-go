package v1

import (
	"strconv"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dtoV1 "onepractice-golang/internal/dto/v1"

	"github.com/gin-gonic/gin"
)

func pathInt(c *gin.Context, name string) (int, bool) {
	value, err := strconv.Atoi(c.Param(name))
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "invalid path parameter"))
		c.Abort()
		return 0, false
	}
	return value, true
}

func pathUint(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 0)
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "invalid path parameter"))
		c.Abort()
		return 0, false
	}
	return uint(value), true
}

func pageQuery(c *gin.Context) dtoV1.PageQuery {
	var query dtoV1.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "invalid pagination parameter"))
		c.Abort()
	}
	return query.Normalize()
}

func newPage[T any](items []T, total int64, q dtoV1.PageQuery) dtoV1.Page[T] {
	q = q.Normalize()
	return dtoV1.Page[T]{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}
}
