package ctxutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHTTPRequestContextHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/path?a=1&b=2", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Forwarded-For", "8.8.8.8, 10.0.0.1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0")
	req.Header.Set("Referer", "https://ref.example.com")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Session-ID", "session-header")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-cookie"})

	ctx := SetHTTPRequest(context.Background(), req)

	if got := GetClientIP(ctx); got != "8.8.8.8" {
		t.Fatalf("GetClientIP() expected 8.8.8.8, got %q", got)
	}
	if got := GetUserAgent(ctx); !strings.Contains(got, "Mozilla/5.0") {
		t.Fatalf("GetUserAgent() returned unexpected value: %q", got)
	}
	if got := GetSessionID(ctx); got != "session-cookie" {
		t.Fatalf("GetSessionID() expected cookie value session-cookie, got %q", got)
	}
	if got := GetReferer(ctx); got != "https://ref.example.com" {
		t.Fatalf("GetReferer() expected referer header, got %q", got)
	}
	if got := GetAcceptLanguage(ctx); got != "en-US,en;q=0.9" {
		t.Fatalf("GetAcceptLanguage() expected en-US,en;q=0.9, got %q", got)
	}
	if got := GetContentType(ctx); got != "application/json" {
		t.Fatalf("GetContentType() expected application/json, got %q", got)
	}
	if got := GetHost(ctx); got != "example.com" {
		t.Fatalf("GetHost() expected example.com, got %q", got)
	}
	if got := GetRemoteAddr(ctx); got != "9.9.9.9:12345" {
		t.Fatalf("GetRemoteAddr() expected 9.9.9.9:12345, got %q", got)
	}
	if got := GetRequestMethod(ctx); got != http.MethodGet {
		t.Fatalf("GetRequestMethod() expected GET, got %q", got)
	}
	if got := GetRequestURL(ctx); got != "https://example.com/path?a=1&b=2" {
		t.Fatalf("GetRequestURL() unexpected value %q", got)
	}

	params := GetQueryParams(ctx)
	if params["a"] != "1" || params["b"] != "2" {
		t.Fatalf("GetQueryParams() returned unexpected params: %v", params)
	}

	if !IsAjaxRequest(ctx) {
		t.Fatal("IsAjaxRequest() expected true")
	}
	if !IsSecureRequest(ctx) {
		t.Fatal("IsSecureRequest() expected true for https URL")
	}
	if got := GetRequestURI(ctx); got != req.RequestURI {
		t.Fatalf("GetRequestURI() expected %q, got %q", req.RequestURI, got)
	}
}

func TestGinContextClientIPHelper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Real-IP", "1.1.1.1")
	ginCtx.Request = req

	ctx := WithGinContext(context.Background(), ginCtx)
	if got := GetClientIP(ctx); got != "1.1.1.1" {
		t.Fatalf("GetClientIP() expected X-Real-IP 1.1.1.1, got %q", got)
	}
}

func TestUserAgentParsing(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0"
	info := ParseUserAgent(ua)
	if info.Browser != "Google Chrome" {
		t.Fatalf("ParseUserAgent() expected browser Google Chrome, got %q", info.Browser)
	}
	if info.OS != "Windows" {
		t.Fatalf("ParseUserAgent() expected OS Windows, got %q", info.OS)
	}
	if info.Mobile {
		t.Fatal("ParseUserAgent() expected non-mobile user agent")
	}

	ctx := SetUserAgent(context.Background(), ua)
	parsed := GetParsedUserAgent(ctx)
	if parsed.Browser != "Google Chrome" {
		t.Fatalf("GetParsedUserAgent() expected browser Google Chrome, got %q", parsed.Browser)
	}
}

func TestPrivateIPCheck(t *testing.T) {
	if !isPrivateIP("10.0.0.1") {
		t.Fatal("isPrivateIP() expected true for private IPv4")
	}
	if !isPrivateIP("::1") {
		t.Fatal("isPrivateIP() expected true for loopback IPv6")
	}
	if isPrivateIP("8.8.8.8") {
		t.Fatal("isPrivateIP() expected false for public IPv4")
	}
	if !isPrivateIP("not-an-ip") {
		t.Fatal("isPrivateIP() expected true for invalid IP")
	}
}
