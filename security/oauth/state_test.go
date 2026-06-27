package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStateManagerSignsAndValidatesState(t *testing.T) {
	manager := NewStateManager("0123456789abcdef0123456789abcdef")

	state, err := manager.GenerateState(&StateData{
		Provider: string(ProviderGoogle),
		Action:   "login",
		NextURL:  "/dashboard",
	})
	if err != nil {
		t.Fatalf("generate state: %v", err)
	}

	parsed, err := manager.ParseState(state)
	if err != nil {
		t.Fatalf("parse state: %v", err)
	}
	if parsed.Provider != string(ProviderGoogle) || parsed.NextURL != "/dashboard" || parsed.Nonce == "" {
		t.Fatalf("unexpected parsed state: %#v", parsed)
	}
}

func TestStateManagerRejectsTamperedState(t *testing.T) {
	manager := NewStateManager("0123456789abcdef0123456789abcdef")
	state, err := manager.GenerateState(&StateData{Provider: string(ProviderGitHub)})
	if err != nil {
		t.Fatalf("generate state: %v", err)
	}

	tampered := "x" + state[1:]
	if _, err := manager.ParseState(tampered); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
}

func TestStateManagerRejectsExpiredState(t *testing.T) {
	manager := NewStateManager("0123456789abcdef0123456789abcdef")
	state, err := manager.GenerateState(&StateData{
		Provider:  string(ProviderGoogle),
		Timestamp: time.Now().Add(-10 * time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("generate state: %v", err)
	}

	if _, err := manager.ParseState(state); !errors.Is(err, ErrStateExpired) {
		t.Fatalf("expected ErrStateExpired, got %v", err)
	}
}

func TestNewClientERequiresStateSecret(t *testing.T) {
	_, err := NewClientE(&Config{Providers: map[string]*ProviderConfig{}})
	if err == nil {
		t.Fatal("expected missing state secret error")
	}
}

func TestGetAuthURLRequiresPKCEChallengeWhenEnabled(t *testing.T) {
	client, err := NewClientE(&Config{
		StateSecret: "0123456789abcdef0123456789abcdef",
		EnablePKCE:  true,
		Providers: map[string]*ProviderConfig{
			string(ProviderGoogle): {
				Enabled:     true,
				ClientID:    "client",
				RedirectURL: "https://example.com/callback",
				AuthURL:     "https://accounts.example.com/auth",
				Scopes:      []string{"openid"},
			},
		},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := client.GetAuthURL(ProviderGoogle, "", nil); err == nil {
		t.Fatal("expected missing code_challenge error")
	}
}

type middlewareClient struct {
	validate func(context.Context, Provider, string) (*TokenInfo, error)
	profile  func(context.Context, Provider, string) (*Profile, error)
}

func (m *middlewareClient) GetAuthURL(Provider, string, map[string]string) (string, error) {
	return "", nil
}

func (m *middlewareClient) ExchangeCodeForToken(context.Context, Provider, string, ...string) (*TokenResponse, error) {
	return nil, nil
}

func (m *middlewareClient) RefreshAccessToken(context.Context, Provider, string) (*TokenResponse, error) {
	return nil, nil
}

func (m *middlewareClient) GetUserProfile(ctx context.Context, provider Provider, token string) (*Profile, error) {
	return m.profile(ctx, provider, token)
}

func (m *middlewareClient) ValidateToken(ctx context.Context, provider Provider, token string) (*TokenInfo, error) {
	return m.validate(ctx, provider, token)
}

func (m *middlewareClient) RevokeToken(context.Context, Provider, string) error {
	return nil
}

func TestMiddlewareValidatesTokenAndStoresContext(t *testing.T) {
	client := &middlewareClient{
		validate: func(ctx context.Context, provider Provider, token string) (*TokenInfo, error) {
			if provider != ProviderGoogle || token != "token" {
				return &TokenInfo{Valid: false}, nil
			}
			return &TokenInfo{Valid: true}, nil
		},
		profile: func(ctx context.Context, provider Provider, token string) (*Profile, error) {
			return &Profile{ID: "user-1", Provider: string(provider)}, nil
		},
	}

	handler := Middleware(client, &MiddlewareOptions{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		Provider:    ProviderGoogle,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetTokenFromContext(r.Context()) != "token" {
			t.Fatal("token was not stored in context")
		}
		if profile := GetProfileFromContext(r.Context()); profile == nil || profile.ID != "user-1" {
			t.Fatalf("profile was not stored in context: %#v", profile)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMiddlewareRejectsWrongScheme(t *testing.T) {
	client := &middlewareClient{
		validate: func(ctx context.Context, provider Provider, token string) (*TokenInfo, error) {
			t.Fatal("validator should not be called")
			return nil, nil
		},
		profile: func(ctx context.Context, provider Provider, token string) (*Profile, error) {
			return nil, nil
		},
	}

	handler := Middleware(client, &MiddlewareOptions{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		Provider:    ProviderGoogle,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
