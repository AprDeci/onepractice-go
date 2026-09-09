package router

import (
	"github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
)

func registerDocsRoutes(r *gin.Engine) {
	r.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
		SpecURL:      "/docs/openapi.json",
		SpecFilePath: "./docs/swagger.json",
		Title:        "Onepractice API",
		Theme:        "light",
	}))
}
