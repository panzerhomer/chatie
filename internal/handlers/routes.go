package handlers

import (
	"github.com/gin-gonic/gin"
)

func Routes(userHandler *userHandler) *gin.Engine {
	r := gin.Default()
	// r.Use(
	// 	gin.LoggerWithWriter(gin.DefaultWriter, "/pathsNotToLog/"),
	// 	gin.Recovery(),
	// )

	ag := r.Group("api/")

	ag.POST("/signup", userHandler.Register)
	ag.POST("/login", userHandler.Login)
	ag.POST("/logout", userHandler.Logout)
	ag.POST("/refresh", userHandler.RefreshAuth)
	ag.POST("/reset-password")
	ag.PUT("/change-password")

	ag.Use(AuthUser(userHandler.tokenManager))
	ag.GET("/hello", userHandler.Hello)

	return r
}
