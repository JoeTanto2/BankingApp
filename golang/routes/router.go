package routes

import (
	"banking-app/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {
	router.POST("/register", controllers.CreateUser())
	router.PUT("/update", controllers.UpdateUser())
	router.POST("/login", controllers.Login())
	router.POST("/register-card", controllers.OpenCard())
	router.POST("/deposit", controllers.Deposit())
	router.POST("/transfer", controllers.Transfer())
}
