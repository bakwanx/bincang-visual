package metrics

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeRooms = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_rooms_total",
			Help: "Total number of active rooms",
		},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_websocket_connections",
			Help: "Total number of active WebSocket connections",
		},
	)

	totalParticipants = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_participants",
			Help: "Total number of participants across all rooms",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeRooms)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(totalParticipants)
}

func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(
			c.Method(),
			c.Path(),
		))
		defer timer.ObserveDuration()

		err := c.Next()

		status := c.Response().StatusCode()
		httpRequestsTotal.WithLabelValues(
			c.Method(),
			c.Path(),
			fmt.Sprintf("%d", status),
		).Inc()

		return err
	}
}

func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

func UpdateActiveRooms(count float64) {
	activeRooms.Set(count)
}

func UpdateActiveConnections(count float64) {
	activeConnections.Set(count)
}
func UpdateTotalParticipants(count float64) {
	totalParticipants.Set(count)
}
