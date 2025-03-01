package controller

import (
	"database/sql"
	"net/http"
)

//go:generate handlerfunc -file $GOFILE

type UserController struct {
	DB *sql.DB
}

// @handlerfunc
func (u *UserController) index(r *http.Request, userID int64) (*models.User, error) {
	return models.Users().Find(userID)
}

type newUserRequest struct {
	Name string
}

// @handlerfunc
func (u *UserController) create(r *http.Request, NewUserRequest newUserRequest) (int64, error) {
	return models.Users().Create(NewUserRequest.Name)
}
