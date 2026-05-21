package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRootMux(gatewayHandler http.Handler) http.Handler {
	rootMux := http.NewServeMux()

	rootMux.Handle("/metrics", promhttp.Handler())

	rootMux.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
		})
	})

	rootMux.HandleFunc("/readiness", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"checks": map[string]string{
				"db":    "not_checked",
				"kafka": "not_checked",
			},
		})
	})

	rootMux.Handle("/", gatewayHandler)

	return rootMux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
