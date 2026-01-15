// Package repository stores export jobs for the full application example.
package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ncobase/ncore/examples/full-application/plugin/export/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JobRepository interface {
	Create(ctx context.Context, job *structs.Job) error
	FindByID(ctx context.Context, id string) (*structs.Job, error)
	FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Job, error)
	Update(ctx context.Context, job *structs.Job) error
	Delete(ctx context.Context, id string) error
}

type jobRepository struct {
	collection *mongo.Collection
}

func NewJobRepository(collection *mongo.Collection) (JobRepository, error) {
	if collection == nil {
		return nil, errors.New("collection is nil")
	}

	repo := &jobRepository{collection: collection}
	if err := repo.ensureIndexes(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *jobRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (r *jobRepository) Create(ctx context.Context, job *structs.Job) error {
	_, err := r.collection.InsertOne(ctx, job)
	return err
}

func (r *jobRepository) FindByID(ctx context.Context, id string) (*structs.Job, error) {
	result := &structs.Job{}
	if err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *jobRepository) FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Job, error) {
	if limit <= 0 {
		limit = 20
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"workspace_id": workspaceID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*structs.Job
	for cursor.Next(ctx) {
		job := &structs.Job{}
		if err := cursor.Decode(job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *jobRepository) Update(ctx context.Context, job *structs.Job) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"id": job.ID}, bson.M{"$set": job})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found: %s", job.ID)
	}
	return nil
}

func (r *jobRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("job not found: %s", id)
	}
	return nil
}

type MemoryJobRepository struct {
	jobs map[string]*structs.Job
	mu   sync.RWMutex
}

func NewMemoryJobRepository() JobRepository {
	return &MemoryJobRepository{
		jobs: make(map[string]*structs.Job),
	}
}

func (r *MemoryJobRepository) Create(ctx context.Context, job *structs.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; exists {
		return fmt.Errorf("job already exists: %s", job.ID)
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *MemoryJobRepository) FindByID(ctx context.Context, id string) (*structs.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	jobCopy := *job
	return &jobCopy, nil
}

func (r *MemoryJobRepository) FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var jobs []*structs.Job
	for _, job := range r.jobs {
		if job.WorkspaceID == workspaceID {
			jobCopy := *job
			jobs = append(jobs, &jobCopy)
		}
	}

	if offset >= len(jobs) {
		return []*structs.Job{}, nil
	}

	end := offset + limit
	if end > len(jobs) {
		end = len(jobs)
	}

	return jobs[offset:end], nil
}

func (r *MemoryJobRepository) Update(ctx context.Context, job *structs.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return fmt.Errorf("job not found: %s", job.ID)
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *MemoryJobRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[id]; !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	delete(r.jobs, id)
	return nil
}
