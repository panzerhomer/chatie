package handlers

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	manager "chatie/pkg/auth"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const UserKeyCtx = "userID"

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	CheckUser(ctx context.Context, email string, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	IsEmailUsed(ctx context.Context, email string) bool
}

type userHandler struct {
	userService  UserService
	tokenManager manager.TokenManager // jwt manager
}

func NewUserhandler(userService UserService, tokenManager manager.TokenManager) *userHandler {
	return &userHandler{
		userService:  userService,
		tokenManager: tokenManager,
	}
}

func (u *userHandler) Register(c *gin.Context) {
	var userRegister models.UserRegister
	if err := c.ShouldBindJSON(&userRegister); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := userRegister.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toUser := &models.User{
		Name:       userRegister.Name,
		Lastname:   userRegister.Lastname,
		Patronymic: userRegister.Patronymic,
		Password:   userRegister.Password,
		Email:      userRegister.Email,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	toUser.GenerateUsername() // generate unique username

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	user, err := u.userService.CreateUser(ctx, toUser)
	if err != nil {
		// select {
		// case <-ctx.Done():
		// 	c.JSON(http.StatusRequestTimeout, gin.H{"error": ctx.Err().Error()})
		// 	return
		// default:
		getErrorResponse(c, err)
		return
		// }
	}

	c.JSON(http.StatusOK, user)
}

type tokenResponse struct {
	Access  string `json:"accessToken"`
	Refresh string `json:"refreshToken"`
}

func (u *userHandler) Login(c *gin.Context) {
	var userLogin models.UserLogin
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	user, err := u.userService.CheckUser(ctx, userLogin.Email, userLogin.Password)
	if err != nil {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout request"})
			return
		default:
			getErrorResponse(c, err)
			return
		}
	}

	tokenAccess, err := u.tokenManager.NewJWT(user.ID, time.Minute*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenRefresh, err := u.tokenManager.NewJWT(user.ID, time.Hour*24*7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokens := tokenResponse{
		Access:  tokenAccess,
		Refresh: tokenRefresh,
	}

	c.SetCookie("access_token", tokenAccess, 3600, "/", "localhost", false, true)

	c.JSON(http.StatusOK, tokens)
}

func (u *userHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (u *userHandler) Hello(c *gin.Context) {
	userID := c.GetInt(UserKeyCtx)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"message": apperror.ErrInternal})
	// 	return
	// }

	// userID, ok := userKeyID.(int)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"message": apperror.ErrInternal})
	// 	return
	// }

	user, err := u.userService.GetUserByID(context.Background(), userID)
	if err != nil {
		getErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": user})
}

func (u *userHandler) RefreshAuth(c *gin.Context) {
	// token, err := c.Cookie("refresh_token")
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err})
	// 	c.Abort()
	// 	return
	// }

	// token
}

func getErrorResponse(c *gin.Context, err error) {
	switch err {
	case apperror.ErrInternal:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	case apperror.ErrUserExists:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case apperror.ErrUserInvalidCredentials:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case apperror.ErrUsersNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case apperror.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
