package handlers

import (
	"chatie/internal/apperror"
	manager "chatie/pkg/auth"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AuthUser(tokenManager manager.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrNotAuthorized.Error()})
			c.Abort()
			return
		}

		userIDFromToken, err := tokenManager.Parse(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrNotAuthorized.Error()})
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(userIDFromToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": apperror.ErrInternal.Error()})
			c.Abort()
			return
		}

		c.Set(UserKeyCtx, userID)

		c.Next()
	}
}
