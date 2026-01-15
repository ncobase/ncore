package wrapper

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/03-multi-module/core/user/structs"
	"github.com/ncobase/ncore/extension/types"
)

type UserServiceWrapper struct {
	em types.ManagerInterface
}

func NewUserServiceWrapper(em types.ManagerInterface) *UserServiceWrapper {
	return &UserServiceWrapper{em: em}
}

func (w *UserServiceWrapper) GetUser(ctx context.Context, userID string) (*structs.User, error) {
	svc, err := w.em.GetCrossService("user", "UserService")
	if err != nil {
		return nil, fmt.Errorf("user service not available: %w", err)
	}

	userSvc, ok := svc.(interface {
		GetUser(ctx context.Context, id string) (*structs.User, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid user service type")
	}

	return userSvc.GetUser(ctx, userID)
}
