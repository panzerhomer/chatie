package repository

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"context"
	"errors"
	"log"
	"regexp"

	"github.com/jackc/pgx/v5"
)

type userRepo struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *userRepo {
	return &userRepo{db}
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.email, 
			u.password,
			u.role,
			e.position,
			u.created_at,
			u.updated_at
		FROM 
			users as u
		JOIN 
			employees as e
		ON
			u.email = e.email
		WHERE 
			u.email = $1`

	var user models.User

	row := r.db.QueryRow(ctx, query, email)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Lastname,
		&user.Patronymic,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Tag,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.email, 
			u.password,
			u.role,
			e.position,
			u.created_at,
			u.updated_at
		FROM 
			users as u
		JOIN 
			employees as e
		ON
			u.email = e.email
		WHERE 
			u.username = $1`

	var user models.User

	row := r.db.QueryRow(ctx, query, username)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Lastname,
		&user.Patronymic,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Tag,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.ErrInternal
	}

	return &user, nil
}

func (r *userRepo) InsertUser(ctx context.Context, user *models.User) (int, error) {
	log.Println("InsertUser befire tx", user)
	tx, err := r.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return -1, apperror.ErrInternal
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()

	queryUsers := `
		INSERT INTO 
			users(username, firstname, lastname, patronymic, email, password) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id`

	log.Println("InsertUser after txt", user)

	var userID int
	err = r.db.QueryRow(ctx, queryUsers, user.Username, user.Name, user.Lastname, user.Patronymic, user.Email, user.Password).Scan(&userID)
	if err != nil {
		if isDuplicateError(err) {
			return -1, apperror.ErrUserExists
		}
		return -1, err
	}

	log.Println("InsertUser after txt", user)

	queryEmployees := `
		INSERT INTO 
			employees(email) 
		VALUES ($1)
	`

	_, err = r.db.Exec(ctx, queryEmployees, user.Email)
	if err != nil {
		log.Println("InsertUser errerr after txt", err)

		return -1, err
	}

	return userID, nil
}

func (r *userRepo) DeleteUser(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.email, 
			u.password,
			u.role,
			e.position,
			u.created_at,
			u.updated_at
		FROM 
			users as u
		JOIN 
			employees as e
		ON
			u.email = e.email
		WHERE 
			u.user_id = $1`

	var user models.User

	row := r.db.QueryRow(ctx, query, userID)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Lastname,
		&user.Patronymic,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Tag,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.ErrInternal
	}

	return &user, nil
}

func (r *userRepo) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.email, 
			u.password,
			u.role,
			e.position,
			u.created_at,
			u.updated_at
		FROM 
			users as u
		JOIN 
			employees as e
		ON
			u.email = e.email`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Name,
			&user.Lastname,
			&user.Patronymic,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.Tag,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			log.Println("ffffffff", err)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperror.ErrUsersNotFound
			}
			return nil, apperror.ErrInternal
		}
		user.Sanitaze()
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, apperror.ErrInternal
	}

	return users, nil
}

func isDuplicateError(err error) bool {
	duplicate := regexp.MustCompile(`\(SQLSTATE 23505\)$`)
	return duplicate.MatchString(err.Error())
}
