package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
)

func NewUserRepo(db *pgxpool.Pool, redis *redis.Client) *UserRepo {
	return &UserRepo{
		db:    db,
		redis: redis,
	}
}

type UserRepo struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) (int, error) {
	var userID int
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email,
		user.PasswordHash,
	).Scan(&userID)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
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
	err := r.db.QueryRow(
		ctx,
		`SELECT u.id, u.email, u.password_hash, EXISTS(SELECT 1 FROM admins WHERE user_id = u.id) AS is_admin FROM users u WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &isAdmin)
	if err != nil {
		if err == pgx.ErrNoRows {
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
	err := r.db.QueryRow(
		ctx,
		`SELECT u.id, u.email, u.password_hash, EXISTS(SELECT 1 FROM admins WHERE user_id = u.id) AS is_admin FROM users u WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &isAdmin)
	if err != nil {
		if err == pgx.ErrNoRows {
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
