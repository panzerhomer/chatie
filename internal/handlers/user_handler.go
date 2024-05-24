package handlers

import (
	"chatie/internal/config"
	"chatie/internal/models"
	manager "chatie/pkg/auth"
	"context"
	"log"
	"net/http"
	"strconv"
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
	config       config.Config
}

func NewUserhandler(
	userService UserService,
	tokenManager manager.TokenManager,
	config config.Config,
) *userHandler {
	return &userHandler{
		userService:  userService,
		tokenManager: tokenManager,
		config:       config,
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

	user, err := u.userService.CreateUser(context.Background(), toUser)
	if err != nil {
		log.Println("{sign up}", err)
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

	user, err := u.userService.CheckUser(context.Background(), userLogin.Email, userLogin.Password)
	if err != nil {
		getErrorResponse(c, err)
		return
	}

	tokenAccess, err := u.tokenManager.NewJWT(user.ID, u.config.Auth.AccessTokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenRefresh, err := u.tokenManager.NewJWT(user.ID, u.config.Auth.RefreshTokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokens := tokenResponse{
		Access:  tokenAccess,
		Refresh: tokenRefresh,
	}

	c.SetCookie("access_token", tokenAccess, 3600, "/", u.config.HTTP.Host, false, true)

	c.JSON(http.StatusOK, tokens)
}

func (u *userHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (u *userHandler) Hello(c *gin.Context) {
	userID := c.GetInt(UserKeyCtx)

	user, err := u.userService.GetUserByID(context.Background(), userID)
	if err != nil {
		getErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

type refreshRequest struct {
	Token string `json:"token"`
}

func (u *userHandler) RefreshAuth(c *gin.Context) {
	var refreshReq refreshRequest
	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := u.tokenManager.Parse(refreshReq.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
		return
	}

	userID, err := strconv.Atoi(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newTokenAccess, err := u.tokenManager.NewJWT(userID, u.config.Auth.AccessTokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newTokenRefresh, err := u.tokenManager.NewJWT(userID, u.config.Auth.RefreshTokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokens := tokenResponse{
		Access:  newTokenAccess,
		Refresh: newTokenRefresh,
	}

	c.SetCookie(
		"access_token",
		newTokenAccess,
		int(u.config.Auth.AccessTokenTTL),
		"/",
		u.config.HTTP.Host,
		false,
		true,
	)

	c.JSON(http.StatusOK, tokens)
}

func getErrorResponse(c *gin.Context, err error) {
	switch err {
	// case apperror.ErrInternal:
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// case apperror.ErrUserExists:
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// case apperror.ErrUserInvalidCredentials:
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// case apperror.ErrUsersNotFound:
	// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	// case apperror.ErrUserNotFound:
	// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
