package controller

import (
	"database/sql"
	"net/http"

	"gitlab.snappcloud.io/doctor/pkg/http"
)

type UserController struct {
	DB *sql.DB
}

func NewUserController() *UserController {
	return UserController{DB: di.GetDB()}
}

func (u *UserController) Index(r *http.Request) (http.Result, error) {}
func (u *UserController) New(r *http.Request, NewUserRequest struct {
	Name string
}) (http.Result, error) {
}

func initUserController() {

	index := func(w http.ResponseWriter, r *http.Request) {
		res, err := NewUserController().Index(r)
	}

	_new := func(w http.ResponseWriter, r *http.Request) {
		res, err := NewUserController().New(r, struct {
			Name string
		})
	}

}
