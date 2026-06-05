package gateway

import (
	"net/http"
	"portal-system/internal/observability/health"
	"portal-system/internal/service"
)

func NewRootMux(gatewayHandler http.Handler, readinessChecker health.Checker, userService service.UserService) http.Handler {
	rootMux := http.NewServeMux()

	RegisterObservabilityHandlers(rootMux, readinessChecker)
	RegisterWebhookHandlers(rootMux, userService)

	rootMux.Handle("/", gatewayHandler)

	return rootMux
}
