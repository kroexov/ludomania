package app

import (
	"context"
	"strconv"
	"time"

	"apisrv/pkg/embedlog"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// registerMetrics is a function that initializes a.stat* variables and adds /metrics endpoint to echo.
func (a *App) registerMetrics() {
	statLogEvents := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: a.appName,
		Subsystem: "log",
		Name:      "events_total",
		Help:      "Log events distributions.",
	}, []string{"type"})

	embedlog.SetStatLogEvents(statLogEvents)

	prometheus.MustRegister(statLogEvents)

	// add db conn metrics
	metrics := NewConnectionPoolMetrics(a.appName)
	prometheus.MustRegister(metrics)
	metrics.ObserveRegularly(context.Background(), a.dbc, "default")

	a.echo.Use(httpMetrics(a.appName))
	a.echo.Any("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// httpMetrics is the middleware function that logs duration of responses.
func httpMetrics(appName string) echo.MiddlewareFunc {
	labels := []string{"method", "uri", "code"}

	echoRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Subsystem: "http",
		Name:      "requests_count",
		Help:      "Requests count by method/path/status.",
	}, labels)

	echoDurations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: appName,
		Subsystem: "http",
		Name:      "responses_duration_seconds",
		Help:      "Response time by method/path/status.",
	}, labels)

	prometheus.MustRegister(echoRequests, echoDurations)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}

			metrics := []string{c.Request().Method, c.Path(), strconv.Itoa(c.Response().Status)}

			echoDurations.WithLabelValues(metrics...).Observe(time.Since(start).Seconds())
			echoRequests.WithLabelValues(metrics...).Inc()

			return nil
		}
	}
}
