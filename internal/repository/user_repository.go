package repository

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"context"
	"errors"
	"log"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepo {
	return &userRepo{db}
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
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

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
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

func (r *userRepo) Create(ctx context.Context, user *models.User) (int, error) {
	tx, err := r.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return -1, err
	}
	defer tx.Rollback(context.Background())

	queryUsers := `
		INSERT INTO 
			users(username, firstname, lastname, patronymic, email, password) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id`

	_, err = tx.Prepare(ctx, "InsertIntoUsers", queryUsers)
	if err != nil {
		return -1, err
	}

	var userID int
	err = tx.QueryRow(ctx, "InsertIntoUsers", user.Username, user.Name, user.Lastname, user.Patronymic, user.Email, user.Password).Scan(&userID)
	if err != nil {
		if isDuplicateError(err) {
			return -1, apperror.ErrUserExists
		}
		return -1, err
	}

	queryEmployees := `INSERT INTO employees(email) VALUES ($1)`

	_, err = tx.Prepare(ctx, "InsertIntoEmp", queryEmployees)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(ctx, "InsertIntoEmp", user.Email)
	if err != nil {
		return -1, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, err
	}

	return userID, nil
}

func (r *userRepo) Delete(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) (*models.User, error) {
	query := `UPDATE users SET`
	param := ""

	switch {
	case user.Info != "":
		query += " info = $2 "
		param = user.Info
	case user.Password != "":
		query += " password = $2 "
		param = user.Info
	}

	query += " WHERE user_id = $1"

	_, err := r.db.Exec(ctx, query, user.ID, param)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, user.ID)
}

func (r *userRepo) GetByID(ctx context.Context, userID int) (*models.User, error) {
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

func (r *userRepo) GetAll(ctx context.Context) ([]models.User, error) {
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
