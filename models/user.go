package models

//go:generate projectx -model $file

type User struct {
	ID   int64
	Name string
}
