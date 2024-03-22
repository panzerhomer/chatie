package repository

import (
	"chatie/internal/models"
	"context"
	"database/sql"
)

type UserRepository interface {
	AddUser(ctx context.Context, user *models.User) error
	RemoveUser(ctx context.Context, user *models.User) error
	FindUserById(ctx context.Context, userID string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db}
}

func (repo *userRepo) AddUser(ctx context.Context, user *models.User) error {
	query, err := repo.db.Prepare("INSERT INTO user(id, name) values(?,?)")
	if err != nil {
		return err
	}

	_, err = query.Exec(user.GetID(), user.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (repo *userRepo) RemoveUser(ctx context.Context, user *models.User) error {
	stmt, err := repo.db.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.GetID())
	if err != nil {
		return err
	}

	return nil
}

func (repo *userRepo) FindUserById(ctx context.Context, userID string) (*models.User, error) {
	row := repo.db.QueryRow("SELECT id, name FROM user where id = ? LIMIT 1", userID)

	var user models.User

	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &user, nil
}

func (repo *userRepo) GetAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := repo.db.Query("SELECT id, name FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}
