package models

import (
	"context"
	"testing"
)

func TestUser(t *testing.T) {
	user := User{
		ID:   1,
		Name: "John Doe",
	}
	Users().Debug().Add(context.Background(), &user, nil)
}
