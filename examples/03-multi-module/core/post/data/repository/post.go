// Package repository provides Mongo-backed post storage for the multi-module example.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/03-multi-module/core/post/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostRepository interface {
	Create(ctx context.Context, post *structs.Post) (*structs.Post, error)
	FindByID(ctx context.Context, id string) (*structs.Post, error)
	List(ctx context.Context) ([]*structs.Post, error)
	Update(ctx context.Context, id, title, content string) (*structs.Post, error)
	Delete(ctx context.Context, id string) error
}

type postRepository struct {
	collection *mongo.Collection
}

func NewPostRepository(db *mongo.Database) PostRepository {
	return &postRepository{collection: db.Collection("posts")}
}

func (r *postRepository) Create(ctx context.Context, post *structs.Post) (*structs.Post, error) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	if _, err := r.collection.InsertOne(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (r *postRepository) FindByID(ctx context.Context, id string) (*structs.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID")
	}

	var post structs.Post
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) List(ctx context.Context) ([]*structs.Post, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*structs.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, id, title, content string) (*structs.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID")
	}

	update := bson.M{
		"$set": bson.M{
			"title":      title,
			"content":    content,
			"updated_at": time.Now(),
		},
	}

	var post structs.Post
	if err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update).Decode(&post); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
