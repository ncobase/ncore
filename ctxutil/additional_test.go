package ctxutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
)

func TestFromGinContextAndConfigHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)

	reqCtx := context.WithValue(context.Background(), "trace", "value")
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil).WithContext(reqCtx)
	ginCtx.Request = req

	extracted := FromGinContext(ginCtx)
	if extracted.Value("trace") != "value" {
		t.Fatalf("unexpected context extracted from gin request: %#v", extracted.Value("trace"))
	}

	ctxWithGin := WithGinContext(context.Background(), ginCtx)
	ginCtx.Set("k", "v")
	if got := GetValue(ctxWithGin, "k"); got != "v" {
		t.Fatalf("expected GetValue to read gin value, got %#v", got)
	}

	cfg := &config.Config{AppName: "ncore-test"}
	ctxWithConfig := SetConfig(context.Background(), cfg)
	if got := GetConfig(ctxWithConfig); got != cfg {
		t.Fatalf("expected GetConfig to return context config, got %#v", got)
	}
}

func TestHTTPHelpersAdditionalBranches(t *testing.T) {
	if req := GetHTTPRequest(context.Background()); req != nil {
		t.Fatalf("expected empty context request to be nil, got %#v", req)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	ginReq := httptest.NewRequest(http.MethodGet, "http://example.com/path?x=1", nil)
	ginReq.Header.Set("X-Session-ID", "session-header")
	ginReq.Header.Set("Referrer", "https://referrer.example.com")
	ginReq.Header.Set("X-Forwarded-Proto", "https")
	ginReq.RemoteAddr = "7.7.7.7:8080"
	ginCtx.Request = ginReq

	ctxWithGin := WithGinContext(context.Background(), ginCtx)
	if got := GetHTTPRequest(ctxWithGin); got != ginReq {
		t.Fatalf("expected GetHTTPRequest to fallback to gin request, got %#v", got)
	}

	ctxWithClientInfo := SetClientInfo(context.Background(), "3.3.3.3", "ua-1", "sid-1")
	ip, ua, sid := GetClientInfo(ctxWithClientInfo)
	if ip != "3.3.3.3" || ua != "ua-1" || sid != "sid-1" {
		t.Fatalf("unexpected client info: ip=%q ua=%q sid=%q", ip, ua, sid)
	}

	ctxWithSession := SetSessionID(context.Background(), "session-direct")
	if got := GetSessionID(ctxWithSession); got != "session-direct" {
		t.Fatalf("expected direct session id, got %q", got)
	}

	ctxWithReferer := SetReferer(context.Background(), "https://manual.example.com")
	if got := GetValue(ctxWithReferer, refererKey); got != "https://manual.example.com" {
		t.Fatalf("expected referer value to be stored in context, got %#v", got)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/path?x=1", nil)
	req.Header.Set("A", "1")
	req.Header.Set("X-Session-ID", "session-header")
	req.Header.Set("Referrer", "https://referrer.example.com")
	req.RemoteAddr = "7.7.7.7:8080"
	ctx := SetHTTPRequest(context.Background(), req)

	if got := GetSessionID(ctx); got != "session-header" {
		t.Fatalf("expected session id from header, got %q", got)
	}
	if got := GetReferer(ctx); got != "https://referrer.example.com" {
		t.Fatalf("expected referer fallback from Referrer header, got %q", got)
	}
	if !IsSecureRequest(ctxWithGin) {
		t.Fatal("expected secure request detection from X-Forwarded-Proto header")
	}
	if IsSecureRequest(ctx) {
		t.Fatal("expected non-secure request for plain http request without TLS/proxy header")
	}
	if IsAjaxRequest(context.Background()) {
		t.Fatal("expected non-AJAX result for empty context")
	}

	headers := GetRequestHeaders(ctx)
	if headers["A"] != "1" {
		t.Fatalf("expected request headers map to include custom header, got %#v", headers)
	}
	if got := GetRequestHeaders(context.Background()); len(got) != 0 {
		t.Fatalf("expected empty headers for context without request, got %#v", got)
	}

	if got := getIPFromAddr("8.8.8.8:1234"); got != "8.8.8.8" {
		t.Fatalf("unexpected host:port IP parse result: %q", got)
	}
	if got := getIPFromAddr("8.8.4.4"); got != "8.8.4.4" {
		t.Fatalf("unexpected plain IP parse result: %q", got)
	}
	if got := getIPFromAddr(""); got != "unknown" {
		t.Fatalf("unexpected empty addr parse result: %q", got)
	}
}
