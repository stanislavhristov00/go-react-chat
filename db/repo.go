package db

import "chat-module/models"

type Repository interface {
	CheckUserExists(username, email string) (bool, error)
	AddUser(user models.User) error
	GetUser(usernameOrEmail string) (*models.User, error)
}
