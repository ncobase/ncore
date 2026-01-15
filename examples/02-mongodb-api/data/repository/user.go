// Package repository provides MongoDB-backed user persistence.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents a user entity.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, skip, limit int64) ([]*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *mongo.Database, logger *logger.Logger) UserRepository {
	collection := db.Collection("users")

	// Create unique index on email
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Warn(ctx, "failed to create index on email", "error", err)
	}

	return &userRepository{
		collection: collection,
		logger:     logger,
	}
}

// Create creates a new user.
func (r *userRepository) Create(ctx context.Context, user *User) (*User, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("user with email %s already exists", user.Email)
		}
		r.logger.Error(ctx, "failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info(ctx, "user created", "id", user.ID.Hex())
	return user, nil
}

// FindByID retrieves a user by ID.
func (r *userRepository) FindByID(ctx context.Context, id string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %s", id)
	}

	var user User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error(ctx, "failed to find user", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByEmail retrieves a user by email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error(ctx, "failed to find user by email", "email", email, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// List retrieves a list of users with pagination.
func (r *userRepository) List(ctx context.Context, skip, limit int64) ([]*User, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error(ctx, "failed to list users", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error(ctx, "failed to decode users", "error", err)
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// Update updates an existing user.
func (r *userRepository) Update(ctx context.Context, user *User) (*User, error) {
	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"updated_at": user.UpdatedAt,
		},
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": user.ID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		if mongo.IsDuplicateKeyError(result.Err()) {
			return nil, fmt.Errorf("user with email %s already exists", user.Email)
		}
		r.logger.Error(ctx, "failed to update user", "id", user.ID.Hex(), "error", result.Err())
		return nil, fmt.Errorf("failed to update user: %w", result.Err())
	}

	var updated User
	if err := result.Decode(&updated); err != nil {
		return nil, fmt.Errorf("failed to decode updated user: %w", err)
	}

	r.logger.Info(ctx, "user updated", "id", user.ID.Hex())
	return &updated, nil
}

// Delete deletes a user by ID.
func (r *userRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", id)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		r.logger.Error(ctx, "failed to delete user", "id", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info(ctx, "user deleted", "id", id)
	return nil
}

// Count returns the total number of users.
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		r.logger.Error(ctx, "failed to count users", "error", err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
