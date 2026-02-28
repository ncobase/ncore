package ctxutil

import (
	"context"
	"testing"
)

func TestContextUserAndPermissionHelpers(t *testing.T) {
	ctx := context.Background()
	ctx = SetUserID(ctx, "u-1")
	ctx = SetUsername(ctx, "alice")
	ctx = SetUserEmail(ctx, "alice@example.com")
	ctx = SetUserStatus(ctx, 1)
	ctx = SetUserIsCertified(ctx, true)
	ctx = SetSpaceID(ctx, "space-1")
	ctx = SetToken(ctx, "token-1")
	ctx = SetProvider(ctx, "local")
	ctx = SetProfile(ctx, map[string]string{"k": "v"})
	ctx = SetUserRoles(ctx, []string{"admin", "editor"})
	ctx = SetUserPermissions(ctx, []string{"read", "write"})
	ctx = SetUserIsAdmin(ctx, true)
	ctx = SetUserSpaceIDs(ctx, []string{"space-1", "space-2"})

	if got := GetUserID(ctx); got != "u-1" {
		t.Fatalf("GetUserID() expected u-1, got %q", got)
	}
	if got := GetUsername(ctx); got != "alice" {
		t.Fatalf("GetUsername() expected alice, got %q", got)
	}
	if got := GetUserEmail(ctx); got != "alice@example.com" {
		t.Fatalf("GetUserEmail() expected alice@example.com, got %q", got)
	}
	if got := GetUserStatus(ctx); got != 1 {
		t.Fatalf("GetUserStatus() expected 1, got %d", got)
	}
	if !GetUserIsCertified(ctx) {
		t.Fatal("GetUserIsCertified() expected true")
	}
	if got := GetSpaceID(ctx); got != "space-1" {
		t.Fatalf("GetSpaceID() expected space-1, got %q", got)
	}
	if got := GetToken(ctx); got != "token-1" {
		t.Fatalf("GetToken() expected token-1, got %q", got)
	}
	if got := GetProvider(ctx); got != "local" {
		t.Fatalf("GetProvider() expected local, got %q", got)
	}
	if profile := GetProfile(ctx); profile == nil {
		t.Fatal("GetProfile() expected non-nil profile")
	}

	roles := GetUserRoles(ctx)
	if len(roles) != 2 || roles[0] != "admin" || roles[1] != "editor" {
		t.Fatalf("GetUserRoles() returned unexpected value: %v", roles)
	}
	perms := GetUserPermissions(ctx)
	if len(perms) != 2 || perms[0] != "read" || perms[1] != "write" {
		t.Fatalf("GetUserPermissions() returned unexpected value: %v", perms)
	}
	if !GetUserIsAdmin(ctx) {
		t.Fatal("GetUserIsAdmin() expected true")
	}
	spaceIDs := GetUserSpaceIDs(ctx)
	if len(spaceIDs) != 2 || spaceIDs[0] != "space-1" || spaceIDs[1] != "space-2" {
		t.Fatalf("GetUserSpaceIDs() returned unexpected value: %v", spaceIDs)
	}
}

func TestEnsureTraceID(t *testing.T) {
	ctx := context.Background()

	nextCtx, id1 := EnsureTraceID(ctx)
	if id1 == "" {
		t.Fatal("EnsureTraceID() should generate a non-empty trace ID")
	}

	finalCtx, id2 := EnsureTraceID(nextCtx)
	if id2 != id1 {
		t.Fatalf("EnsureTraceID() should preserve existing trace ID, got %q then %q", id1, id2)
	}

	if got := GetTraceID(finalCtx); got != id1 {
		t.Fatalf("GetTraceID() expected %q, got %q", id1, got)
	}
}

func TestExtractContext(t *testing.T) {
	baseCtx := context.WithValue(context.Background(), "k", "v")
	payload := map[string]any{
		"ctx":   baseCtx,
		"other": "x",
	}

	got := ExtractContext(&payload)
	if got != baseCtx {
		t.Fatal("ExtractContext() should return embedded context")
	}
	if _, exists := payload["ctx"]; exists {
		t.Fatal("ExtractContext() should delete ctx key from payload")
	}

	if got := ExtractContext(nil); got == nil {
		t.Fatal("ExtractContext(nil) should return background context")
	}
}
