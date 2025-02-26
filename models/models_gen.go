package models

type UserQueryBuilder struct{}

func (u *UserQueryBuilder) Find(userID int64) (*User, error)  {}
func (u *UserQueryBuilder) Create(name string) (int64, error) {}

func Users() *UserQueryBuilder {}
