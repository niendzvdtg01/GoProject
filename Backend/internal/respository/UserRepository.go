package respository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user model.User) error {
	const query = `
	INSERT INTO users (user_id, username, email, password_hash, role)
	VALUES (?, ?, ?, ?, ?);`

	_, err := r.db.ExecContext(ctx, query, user.UserID, user.Username, user.Email, user.PasswordHash, user.Role)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	const query = `
	SELECT user_id, username, email, password_hash, role, created_at
	FROM users
	WHERE email = ?;`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]model.PublicUser, error) {
	const query = `
	SELECT user_id, username, email, created_at
	FROM users
	ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]model.PublicUser, 0)
	for rows.Next() {
		var user model.PublicUser
		if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetUserByUsername(username string) (model.User, error) {

	const query = `
	SELECT user_id, username, email, password_hash, role, created_at
	FROM users
	WHERE username = ?;`

	var user model.User
	err := r.db.QueryRow(query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(userID string) (model.User, error) {
	const query = `
	SELECT user_id, username, email, password_hash, role, created_at
	FROM users
	WHERE user_id = ?;`

	var user model.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}
