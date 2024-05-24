package handlers

import (
	"chatie/internal/handlers/middleware"
	"chatie/internal/ws"

	"github.com/gin-gonic/gin"
)

func Routes(userHandler *userHandler, hub *ws.WsServer) *gin.Engine {
	r := gin.Default()

	ag := r.Group("api/")

	ag.POST("/signup", userHandler.Register)
	ag.POST("/login", userHandler.Login)
	ag.POST("/logout", userHandler.Logout)
	ag.POST("/refresh", userHandler.RefreshAuth)
	// ag.POST("/reset-password")
	// ag.PUT("/change-password")

	ag.Use(middleware.AuthUser(userHandler.tokenManager))
	ag.GET("/user/me", userHandler.Hello)
	ag.GET("/user/ws", func(c *gin.Context) {
		ws.ServeWS(hub, c)
	})

	return r
}
