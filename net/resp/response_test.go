package resp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSuccessSetsContentTypeBeforeStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	Success(rec, map[string]string{"status": "ok"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("unexpected content type: %q", contentType)
	}
}

func TestFailAbortDoesNotDoubleWrite(t *testing.T) {
	rec := httptest.NewRecorder()

	Fail(rec, BadRequest("bad request"), true)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if body := rec.Body.String(); strings.Count(body, "\n") != 1 {
		t.Fatalf("unexpected response body, possible double write: %q", body)
	}
}
