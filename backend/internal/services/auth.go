package services

import (
	"log"
	"spotiscan/pkg/db"
)

func NewAuthService(db *db.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

type AuthService struct {
	db *db.DB
}

func (s *AuthService) Signup(username, email, password string) (string, error) {
	emailUsed, err := s.db.EmailUsed(email)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	if emailUsed {
		log.Println(err)
		return "", ErrEmailUsed
	}

	usernameUsed, err := s.db.UsernameExists(username)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	if usernameUsed {
		log.Println(err)
		return "", ErrUsernameUsed
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Println(err)
		return "", ErrInternal
	}

	token := generateRandomString()

	err = s.db.CreateUserWithSession(username, email, passwordHash, token)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}

	return token, nil
}

func (s *AuthService) Logout(sessionToken string) error {
	err := s.db.DeleteSession(sessionToken)
	if err != nil {
		log.Println(err)
		return ErrDatabaseFailure
	}

	return nil
}

func (s *AuthService) Login(emailOrUsername, password string) (string, error) {
	userId, err := s.db.GetUserIDByEmailOrUsername(emailOrUsername)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	if userId == 0 {
		return "", ErrInvalidCredentials
	}

	passwordHash, err := s.db.GetPasswordHashByUserID(userId)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}

	if !passwordMatchesHash(password, passwordHash) {
		return "", ErrInvalidCredentials
	}

	token := generateRandomString()

	err = s.db.CreateSession(userId, token)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}

	return token, nil
}
