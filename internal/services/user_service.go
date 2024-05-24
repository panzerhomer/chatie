package services

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (int, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, email string) (*models.User, error)
	Delete(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID int) (*models.User, error)
	GetAll(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
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

	userID, err := u.userRepo.Create(ctx, user)
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

func (u *userService) UpdateInfo(ctx context.Context, user *models.User) (*models.User, error) {
	user, err := u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrUserNotUpdated
	}

	return user, nil
}

func (u *userService) IsEmailUsed(ctx context.Context, email string) bool {
	_, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return true
	}

	return false
}

func (u *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return u.userRepo.GetAll(ctx)
}
