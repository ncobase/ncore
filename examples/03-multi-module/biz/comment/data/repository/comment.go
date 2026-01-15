package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *structs.Comment) (*structs.Comment, error)
	FindByID(ctx context.Context, id string) (*structs.Comment, error)
	ListByPost(ctx context.Context, postID string) ([]*structs.Comment, error)
	Delete(ctx context.Context, id string) error
}

type commentRepository struct {
	collection *mongo.Collection
}

func NewCommentRepository(db *mongo.Database) CommentRepository {
	return &commentRepository{collection: db.Collection("comments")}
}

func (r *commentRepository) Create(ctx context.Context, comment *structs.Comment) (*structs.Comment, error) {
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	if _, err := r.collection.InsertOne(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *commentRepository) FindByID(ctx context.Context, id string) (*structs.Comment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid comment ID")
	}

	var comment structs.Comment
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *commentRepository) ListByPost(ctx context.Context, postID string) ([]*structs.Comment, error) {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"post_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*structs.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
