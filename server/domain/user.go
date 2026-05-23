package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password,omitempty" bson:"password"`
	Role     string             `json:"role,omitempty" bson:"role"`
	Pages    []string           `json:"pages,omitempty" bson:"pages"`
}

type UserRepository interface {
	Save(ctx context.Context, email, password string) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}
