package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"portal-system/internal/observability/health"
)

func TestReadinessReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	t.Parallel()

	handler := NewMux(func(context.Context) health.Report {
		return health.NewReport(map[string]string{
			"db":    health.StatusOK,
			"kafka": health.StatusOK,
			"smtp":  health.StatusError,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"status":"not_ready"`) {
		t.Fatalf("expected not_ready body, got %s", body)
	}
}

func TestReadinessReturnsOKWhenDependenciesAreHealthy(t *testing.T) {
	t.Parallel()

	handler := NewMux(func(context.Context) health.Report {
		return health.NewReport(map[string]string{
			"db":    health.StatusOK,
			"kafka": health.StatusOK,
			"smtp":  health.StatusOK,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
