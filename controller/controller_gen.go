package controller

import "net/http"

func (u *UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /create", func(w http.ResponseWriter, r *http.Request) {
		var input1 newUserRequest

		u.create(r, input1)
	})
	///
	mux.ServeHTTP(w, r)
}
