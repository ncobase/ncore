package ctxutil

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/messaging/email"
	"github.com/ncobase/ncore/oss"
)

type mockSender struct {
	messageID string
	err       error
}

func (m *mockSender) SendTemplateEmail(_ string, _ email.Template) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.messageID, nil
}

type mockStorage struct{}

func (m *mockStorage) Get(_ string) (*os.File, error) { return nil, errors.New("not implemented") }
func (m *mockStorage) GetStream(_ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) Put(_ string, _ io.Reader) (*oss.Object, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) Delete(_ string) error                { return nil }
func (m *mockStorage) List(_ string) ([]*oss.Object, error) { return nil, nil }
func (m *mockStorage) GetURL(_ string) (string, error)      { return "", nil }
func (m *mockStorage) GetEndpoint() string                  { return "mock-endpoint" }
func (m *mockStorage) Exists(_ string) (bool, error)        { return false, nil }
func (m *mockStorage) Stat(_ string) (*oss.Object, error)   { return nil, errors.New("not implemented") }

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

func TestEmailHelpers(t *testing.T) {
	sender := &mockSender{messageID: "message-1"}
	ctx := SetEmailSender(context.Background(), sender)

	gotSender, err := GetEmailSender(ctx)
	if err != nil {
		t.Fatalf("GetEmailSender returned unexpected error: %v", err)
	}
	if gotSender != sender {
		t.Fatalf("GetEmailSender returned unexpected sender instance: %#v", gotSender)
	}

	messageID, err := SendEmailWithTemplate(ctx, "a@example.com", email.Template{Subject: "s"})
	if err != nil {
		t.Fatalf("SendEmailWithTemplate returned unexpected error: %v", err)
	}
	if messageID != "message-1" {
		t.Fatalf("unexpected message ID: %q", messageID)
	}

	ctxUnknown := SetConfig(context.Background(), &config.Config{
		Email: &config.Email{Provider: "unknown-provider"},
	})
	if _, err := GetEmailSender(ctxUnknown); err == nil {
		t.Fatal("expected unknown provider error")
	}

	ctxSendErr := SetEmailSender(context.Background(), &mockSender{err: errors.New("send failure")})
	if _, err := SendEmailWithTemplate(ctxSendErr, "a@example.com", email.Template{Subject: "s"}); err == nil {
		t.Fatal("expected send error")
	}
}

func TestStorageHelpers(t *testing.T) {
	cfg := &config.Config{
		Storage: &oss.Config{
			Provider: "filesystem",
			Bucket:   t.TempDir(),
		},
	}

	mock := &mockStorage{}
	ctx := SetStorage(SetConfig(context.Background(), cfg), mock)
	gotStorage, gotCfg := GetStorage(ctx)
	if gotStorage == nil {
		t.Fatal("expected GetStorage to return mock storage")
	}
	if gotStorage != mock {
		t.Fatalf("expected same storage instance, got %#v", gotStorage)
	}
	if gotCfg == nil || gotCfg.Provider != "filesystem" {
		t.Fatalf("unexpected storage config: %#v", gotCfg)
	}

	ctxAuto := SetConfig(context.Background(), &config.Config{
		Storage: &oss.Config{Provider: "filesystem", Bucket: t.TempDir()},
	})
	autoStorage, autoCfg := GetStorage(ctxAuto)
	if autoStorage == nil {
		t.Fatal("expected auto storage creation to succeed")
	}
	if autoCfg == nil || autoCfg.Provider != "filesystem" {
		t.Fatalf("unexpected auto storage config: %#v", autoCfg)
	}

	ctxInvalid := SetConfig(context.Background(), &config.Config{Storage: nil})
	invalidStorage, invalidCfg := GetStorage(ctxInvalid)
	if invalidStorage != nil || invalidCfg != nil {
		t.Fatalf("expected invalid storage config to return nil values, got storage=%#v cfg=%#v", invalidStorage, invalidCfg)
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
