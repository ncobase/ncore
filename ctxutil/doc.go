// Package ctxutil provides context utilities for managing request-scoped values
// and operations in web applications.
//
// This package offers helpers for:
//   - Storing and retrieving user information (ID, username, profile)
//   - Managing Gin context integration
//   - Handling storage, email, and SMS services via context
//   - Generating business codes and tracking request IDs
//   - Async operations with timeout management
//
// # Context Value Management
//
// Store and retrieve values from context:
//
//	ctx := ctxutil.SetUserID(ctx, "user-123")
//	userID := ctxutil.GetUserID(ctx)
//
//	ctx = ctxutil.SetUsername(ctx, "john.doe")
//	username := ctxutil.GetUsername(ctx)
//
// # Gin Integration
//
// Extract Gin context from standard context:
//
//	ginCtx := ctxutil.GetGinContext(ctx)
//	if ginCtx != nil {
//	    ginCtx.JSON(200, data)
//	}
//
// # Business Code Generation
//
// Generate unique business tracking codes:
//
//	code := ctxutil.GenerateBusinessCode("ORD") // e.g., "ORD202602ABC0001"
//
// # Async Operations
//
// Execute operations asynchronously with automatic timeout:
//
//	ctxutil.AsyncRun(ctx, func(ctx context.Context) error {
//	    // Your async work here
//	    return doWork(ctx)
//	})
//
// All async operations respect context cancellation and timeouts.
package ctxutil
