package gateway

import (
	"encoding/json"
	"net/http"
	"portal-system/internal/observability/health"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterObservabilityHandlers(mux *http.ServeMux, readinessChecker health.Checker) {
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
		})
	})

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		report := health.NewReport(map[string]string{})
		if readinessChecker != nil {
			report = readinessChecker(r.Context())
		}

		writeJSON(w, health.HTTPStatus(report), report)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
