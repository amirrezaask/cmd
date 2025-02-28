package controller

import (
	"database/sql"
	"net/http"
)

func _() {
	http.Handle("/users", &UserController{})
}

type UserController struct {
	DB *sql.DB
}

// @httpHandlerFunc
func (u *UserController) index(r *http.Request, userID int64) (*models.User, error) {
	return models.Users().Find(userID)
}

type newUserRequest struct {
	Name string
}

func (u *UserController) create(r *http.Request, NewUserRequest newUserRequest) (int64, error) {
	return models.Users().Create(NewUserRequest.Name)
}

// generated
func (u *UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
