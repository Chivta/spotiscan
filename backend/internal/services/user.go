package services

import (
	"log"
	"spotiscan/pkg/db"
	"spotiscan/models"
)

func NewUserService(db *db.DB) *UserService {
	return &UserService{
		db: db,
	}
}

type UserService struct {
	db *db.DB
}


func (s *UserService) GetUser(id int) (*models.User, error){
	user,err := s.db.GetUserByID(id)
	if err != nil {
		log.Println(err)
		return nil, ErrDatabaseFailure
	}

	return user, nil	
} 