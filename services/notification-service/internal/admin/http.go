package admin

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMux() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
		})
	})

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"checks": map[string]string{
				"db":    "not_checked",
				"kafka": "not_checked",
				"smtp":  "not_checked",
			},
		})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
