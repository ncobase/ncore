package wrapper

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/03-multi-module/core/post/structs"
	"github.com/ncobase/ncore/extension/types"
)

type PostServiceWrapper struct {
	em types.ManagerInterface
}

func NewPostServiceWrapper(em types.ManagerInterface) *PostServiceWrapper {
	return &PostServiceWrapper{em: em}
}

func (w *PostServiceWrapper) GetPost(ctx context.Context, postID string) (*structs.Post, error) {
	svc, err := w.em.GetCrossService("post", "PostService")
	if err != nil {
		return nil, fmt.Errorf("post service not available: %w", err)
	}

	postSvc, ok := svc.(interface {
		GetPost(ctx context.Context, id string) (*structs.Post, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid post service type")
	}

	return postSvc.GetPost(ctx, postID)
}
