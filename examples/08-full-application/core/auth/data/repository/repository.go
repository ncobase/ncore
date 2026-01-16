package repository

import (
	"context"

	"github.com/ncobase/ncore/examples/08-full-application/core/auth/structs"
)

type UserRepository interface {
	Create(ctx context.Context, user *structs.User) error
	FindByID(ctx context.Context, id string) (*structs.User, error)
	FindByEmail(ctx context.Context, email string) (*structs.User, error)
	Update(ctx context.Context, user *structs.User) error
	Delete(ctx context.Context, id string) error
}
