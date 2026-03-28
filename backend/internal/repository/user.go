package repository

import (
	"context"
	"database/sql"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/lib/pq"

	"github.com/redis/go-redis/v9"

)

func NewUserRepo(logger *logger.Logger, db *sql.DB, redis *redis.Client) *UserRepo {
	return &UserRepo{
		logger:  logger,
		db:      db,
		redis:   redis,
	}
}

type UserRepo struct {
	logger  *logger.Logger
	db      *sql.DB
	redis   *redis.Client
}

func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) (int, error) {
	var userID int
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email,
		user.PasswordHash,
	).Scan(&userID)

	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, appErrors.ErrEmailExists
		}
		return 0, appErrors.ErrDatabaseFailure
	}
	return userID, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return nil, appErrors.ErrNotFound
		}
		return nil, appErrors.ErrDatabaseFailure
	}
	user.Role = models.RoleUser
	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return nil, appErrors.ErrUnauthorized
		}
		return nil, appErrors.ErrDatabaseFailure
	}
	user.Role = models.RoleUser
	return &user, nil
}
