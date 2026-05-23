package repository

import (
	"context"
	"fmt"
	"os"
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

func (repo *UserRepository) EnsureIndexes(ctx context.Context) error {
	collection := repo.db.Collection("users")
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (repo *UserRepository) Save(ctx context.Context, email, password string) error {
	collection := repo.db.Collection("users")
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	_, err := collection.InsertOne(ctx, domain.User{
		Email:    normalizedEmail,
		Password: password,
		Role:     "user",
		Pages:    []string{"reports"},
	})
	return err
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	collection := repo.db.Collection("users")
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	var user domain.User
	err := collection.FindOne(ctx, bson.M{"email": normalizedEmail}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) EnsureDefaultAdmin(ctx context.Context) error {
	collection := repo.db.Collection("users")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"email":    "admin@test.com",
			"password": string(hashedPassword),
			"role":     "admin",
			"pages":    []string{"photos", "devices", "statistics", "reports", "users"},
		},
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"email": "admin@test.com"},
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (repo *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	collection := repo.db.Collection("users")
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"email": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := []domain.User{}
	for cursor.Next(ctx) {
		var user domain.User
		if decodeErr := cursor.Decode(&user); decodeErr != nil {
			return nil, decodeErr
		}
		users = append(users, user)
	}
	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}
	return users, nil
}

func (repo *UserRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	collection := repo.db.Collection("users")
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid inserted id type")
	}
	user.ID = objectID
	return &user, nil
}

func (repo *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	collection := repo.db.Collection("users")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user domain.User
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) UpdateByID(ctx context.Context, id string, setFields map[string]any) error {
	collection := repo.db.Collection("users")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	if len(setFields) == 0 {
		return nil
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": setFields})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *UserRepository) DeleteByID(ctx context.Context, id string) error {
	collection := repo.db.Collection("users")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
