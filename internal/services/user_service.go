package services

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"context"
)

type UserRepository interface {
	InsertUser(ctx context.Context, user *models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, email string) (*models.User, error)
	DeleteUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *userService {
	return &userService{userRepo: userRepo}
}

func (u *userService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	user.Password = hashedPassword

	userID, err := u.userRepo.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = userID

	return user, nil
}

func (u *userService) CheckUser(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := u.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := checkPassword(password, user.Password); err != nil {
		return nil, err
	}

	user.Password = ""

	return user, nil
}

func (u *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user.Password = ""

	return user, nil
}

func (u *userService) IsEmailUsed(ctx context.Context, email string) bool {
	_, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return true
	}

	return false
}

func (u *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return u.userRepo.GetAllUsers(ctx)
	// return user, err
}
