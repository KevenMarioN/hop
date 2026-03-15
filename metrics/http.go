package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns an HTTP handler that exposes registered metrics
// in the Prometheus text format.
func Handler(registry prometheus.Registerer) http.Handler {
	return promhttp.Handler()
}
