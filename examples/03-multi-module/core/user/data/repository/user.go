// Package repository provides Mongo-backed user storage for the multi-module example.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/03-multi-module/core/user/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(ctx context.Context, user *structs.User) (*structs.User, error)
	FindByID(ctx context.Context, id string) (*structs.User, error)
	List(ctx context.Context) ([]*structs.User, error)
	Update(ctx context.Context, id, name, email string) (*structs.User, error)
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{collection: db.Collection("users")}
}

func (r *userRepository) Create(ctx context.Context, user *structs.User) (*structs.User, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if _, err := r.collection.InsertOne(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*structs.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	var user structs.User
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) List(ctx context.Context) ([]*structs.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*structs.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id, name, email string) (*structs.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	update := bson.M{
		"$set": bson.M{
			"name":       name,
			"email":      email,
			"updated_at": time.Now(),
		},
	}

	var user structs.User
	if err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
