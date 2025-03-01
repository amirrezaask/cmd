package models

//go:generate modelgen -file $GOFILE

// @querybuilder
type User struct {
	ID   int64
	Name string
}
