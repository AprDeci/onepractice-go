package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, Msg: "success", Data: data})
}

func SuccessNoData(c *gin.Context) {
	Success(c, nil)
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Body{Code: code, Msg: msg, Data: nil})
}
