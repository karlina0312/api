package controllers

import (
	"net/http"

	gin "github.com/gin-gonic/gin"
	databases "gitlab.com/fibocloud/aws-billing/api_v2/databases"
	middlewares "gitlab.com/fibocloud/aws-billing/api_v2/middlewares"
	structs "gitlab.com/fibocloud/aws-billing/api_v2/structs"
)

// Init Controller
func Init(router *gin.RouterGroup) {
	db := databases.InitDB()
	bc := BaseController{
		Response: &structs.Response{
			StatusCode: http.StatusOK,
			Body: structs.ResponseBody{
				StatusCode: 0,
				ErrorMsg:   "",
				Body:       nil,
			},
		},
		DB: db,
	}
	AuthController{bc}.Init(router.Group("/auth"))
	authRouter := router.Group("")
	authRouter.Use(middlewares.Authenticate(db))

	{
		UserController{bc}.Init(authRouter.Group("/user"))
		ConstExplorerController{bc}.Init(authRouter.Group("/aws"))
		CredentialsController{bc}.Init(authRouter.Group("/credentials"))
	}
}
