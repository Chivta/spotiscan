package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"

	"github.com/lib/pq"

	"github.com/redis/go-redis/v9"
)

func NewUserRepo(db *sql.DB, redis *redis.Client) *UserRepo {
	return &UserRepo{
		db:    db,
		redis: redis,
	}
}

type UserRepo struct {
	db    *sql.DB
	redis *redis.Client
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) (int, error) {
	var userID int
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email,
		user.PasswordHash,
	).Scan(&userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, domain.ErrEmailExists
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return userID, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	var isAdmin bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.email, u.password_hash, EXISTS(SELECT 1 FROM admins WHERE user_id = u.id) AS is_admin FROM users u WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	if isAdmin {
		user.Role = domain.RoleAdmin
	} else {
		user.Role = domain.RoleUser
	}
	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	var isAdmin bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.email, u.password_hash, EXISTS(SELECT 1 FROM admins WHERE user_id = u.id) AS is_admin FROM users u WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUnauthorized
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	if isAdmin {
		user.Role = domain.RoleAdmin
	} else {
		user.Role = domain.RoleUser
	}
	return &user, nil
}
