package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"spotiscan/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

func NewUserService(db *db.DB) *UserService {
	return &UserService{
		db: db,
	}
}

type UserService struct {
	db *db.DB
}

var (
	ErrInternal        = errors.New("general internal error")
	ErrUserExists      = errors.New("user already exists")
	ErrEmailUsed       = errors.New("email already in use")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrDatabaseFailure = errors.New("database operation failed")
)


func hashPassword(password string) (string, error) {
	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(password_hash), nil
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *UserService) Signup(username, email, password string) (string, error) {
	emailUsed, err := s.db.EmailUsed(email)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	if emailUsed {
		log.Println(err)
		return "", ErrEmailUsed
	}

	usernameUsed, err := s.db.UsernameExists(email)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	if usernameUsed { // TODO: doesnt seem to work
		log.Println(err)
		return "", ErrUserExists
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Println(err)
		return "", ErrInternal
	}

	token := generateSessionToken()

	err = s.db.CreateUserWithSession(username, email, passwordHash,token)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}

	log.Println(token)

	return token, nil
}
