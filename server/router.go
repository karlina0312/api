package server

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/fibocloud/aws-billing/api_v2/controllers"
)

// Routers ...
func Routers(app *gin.Engine) *gin.Engine {
	api := app.Group("/api/v2")
	controllers.Init(api)
	return app
}
