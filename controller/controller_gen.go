package controller

import "net/http"


func (u *UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	///
	mux.ServeHTTP(w, r)
}