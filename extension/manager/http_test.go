package manager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extensionconfig "github.com/ncobase/ncore/extension/config"
	"github.com/ncobase/ncore/extension/types"
)

func testManager() *Manager {
	return &Manager{
		extensions: make(map[string]*types.Wrapper),
		conf: &config.Config{
			Extension: &extensionconfig.Config{
				HotReload:  true,
				MaxPlugins: 10,
			},
		},
		startTime: time.Now().Add(-time.Minute),
	}
}

func TestManageRoutesWithoutGuardOnlyMountsHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	manager := testManager()
	manager.ManageRoutes(router.Group("/ncore"))

	healthReq := httptest.NewRequest(http.MethodGet, "/ncore/health", nil)
	healthRec := httptest.NewRecorder()
	router.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected health route to be public, got %d", healthRec.Code)
	}

	systemReq := httptest.NewRequest(http.MethodGet, "/ncore/system/info", nil)
	systemRec := httptest.NewRecorder()
	router.ServeHTTP(systemRec, systemReq)
	if systemRec.Code != http.StatusNotFound {
		t.Fatalf("expected system route to require guard mounting, got %d", systemRec.Code)
	}
}

func TestManageRoutesWithGuardMountsSystemInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	manager := testManager()
	manager.ManageRoutesWithGuard(router.Group("/ncore"), func(c *gin.Context) {
		c.Header("x-guarded", "true")
		c.Next()
	})

	req := httptest.NewRequest(http.MethodGet, "/ncore/system/info", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected guarded system route, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("x-guarded") != "true" {
		t.Fatal("guard middleware was not applied")
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["started_at"] == "" || body["uptime"] == "" {
		t.Fatalf("expected runtime info, got %#v", body)
	}
	if _, ok := body["version"]; ok {
		t.Fatalf("system info should not expose hard-coded version: %#v", body)
	}
}
