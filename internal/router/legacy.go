package router

import "github.com/gin-gonic/gin"

func registerLegacyPublicRoutes(rg *gin.RouterGroup, deps Deps) {
	rg.POST("/user/register", deps.LegacyUser.Register)
	rg.POST("/user/login", deps.LegacyUser.Login)
	rg.POST("/user/resetpassword", deps.LegacyUser.ResetPassword)

	rg.GET("/captcha/email", deps.LegacyCaptcha.Email)
	rg.POST("/captcha/email/verify", deps.LegacyCaptcha.VerifyEmail)

	rg.GET("/paper/all", deps.LegacyPaper.All)
	rg.POST("/paper/getPaperwithQuerys", deps.LegacyPaper.Page)
	rg.POST("/paper/getPaperandRatingWithQuerys", deps.LegacyPaper.PageWithRating)
	rg.GET("/paper/type", deps.LegacyPaper.ByType)
	rg.GET("/paper/types", deps.LegacyPaper.Types)
	rg.GET("/paper/intro", deps.LegacyPaper.Intro)

	rg.GET("/question/getById", deps.LegacyQuestion.ByPaperID)
	rg.GET("/question/getByType", deps.LegacyQuestion.ByPaperIDAndType)
	rg.GET("/question/getAllByIdSplitByPart", deps.LegacyQuestion.SplitByPart)
	rg.GET("/question/getAnswersByPaperId", deps.LegacyQuestion.Answers)
	rg.POST("/question/practice", deps.LegacyQuestion.Practice)

	rg.GET("/dictionary/lookup", deps.LegacyDictionary.LookupMeanings)
	rg.GET("/dictionary/words", deps.LegacyDictionary.ListWords)
	rg.GET("/dictionary/words/:wordid", deps.LegacyDictionary.GetWordDetail)
	rg.GET("/dictionary/words/spelling/:spelling", deps.LegacyDictionary.GetWordBySpelling)
	rg.GET("/dictionary/books", deps.LegacyDictionary.ListBooks)
	rg.GET("/dictionary/books/:bookid/words", deps.LegacyDictionary.ListBookWords)
}

func registerLegacyProtectedRoutes(rg *gin.RouterGroup, deps Deps) {
	rg.GET("/user/info", deps.LegacyUser.Info)
	rg.POST("/user/logout", deps.LegacyUser.Logout)
	rg.POST("/record/save", deps.LegacyRecord.Save)
	rg.GET("/record/list", deps.LegacyRecord.List)
	rg.POST("/record/update", deps.LegacyRecord.Update)
	rg.POST("/word/favorites", deps.LegacyWordFavorite.Add)
	rg.DELETE("/word/favorites", deps.LegacyWordFavorite.Remove)
	rg.GET("/word/favorites/check", deps.LegacyWordFavorite.Check)
	rg.GET("/word/favorites", deps.LegacyWordFavorite.List)
}
