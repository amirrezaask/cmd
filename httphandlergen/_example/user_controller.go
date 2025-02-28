package controller

import (
	"database/sql"
	"net/http"
	"x/models"
)

type UserController struct {
	DB *sql.DB
}

func (u *UserController) index(r *http.Request, userID int64) (*models.User, error) {
	return models.Users().Find(userID)
}

type newUserRequest struct {
	Name string
}

func (u *UserController) create(r *http.Request, NewUserRequest newUserRequest) (int64, error) {
	return models.Users().Create(NewUserRequest.Name)
}
