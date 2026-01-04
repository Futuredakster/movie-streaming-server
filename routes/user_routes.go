package routes

import (
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/register", controllers.RegisterUser())
	router.POST("/login", controllers.Login())
}
