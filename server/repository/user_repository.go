package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"mqtt-streaming-server/domain"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Save(ctx context.Context, email, password string) error {
	collection := repo.db.Collection("users")
	_, err := collection.InsertOne(ctx, domain.User{
		Email:    email,
		Password: password,
		Role:     "user",
		Pages:    []string{"reports"},
	})
	return err
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	collection := repo.db.Collection("users")
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user domain.User
		if decodeErr := cursor.Decode(&user); decodeErr != nil {
			return nil, decodeErr
		}
		if strings.EqualFold(strings.TrimSpace(user.Email), normalizedEmail) {
			return &user, nil
		}
	}
	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}
	return nil, mongo.ErrNoDocuments
}

func (repo *UserRepository) EnsureDefaultAdmin(ctx context.Context) error {
	collection := repo.db.Collection("users")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	adminPages := []string{"photos", "devices", "statistics", "reports", "users"}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"email": "admin@test.com"},
		bson.M{
			"$set": bson.M{
				"email":    "admin@test.com",
				"password": string(hashedPassword),
				"role":     "admin",
				"pages":    adminPages,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to ensure default admin: %w", err)
	}
	return nil
}

func (repo *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	collection := repo.db.Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UserRepository) Create(ctx context.Context, user domain.User) error {
	collection := repo.db.Collection("users")
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	_, err := collection.InsertOne(ctx, user)
	return err
}

func (repo *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	collection := repo.db.Collection("users")
	var user domain.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) UpdateByID(ctx context.Context, id string, update map[string]any) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user id")
	}

	collection := repo.db.Collection("users")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	return err
}

func (repo *UserRepository) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user id")
	}

	collection := repo.db.Collection("users")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
