package middleware

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmkteam/zenrpc/v2"
)

// WithMetrics logs duration of RPC requests via Prometheus. Default AppName is zenrpc. It exposes two
// metrics: `appName_rpc_error_requests_count` and `appName_rpc_responses_duration_seconds`. Labels: method, code,
// platform, version.
func WithMetrics(appName string) zenrpc.MiddlewareFunc {
	if appName == "" {
		appName = "zenrpc"
	}

	rpcErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Subsystem: "rpc",
		Name:      "error_requests_count",
		Help:      "Error requests count by method and error code.",
	}, []string{"method", "code", "platform", "version"})

	rpcDurations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: appName,
		Subsystem: "rpc",
		Name:      "responses_duration_seconds",
		Help:      "Response time by method and error code.",
	}, []string{"method", "code", "platform", "version"})

	prometheus.MustRegister(rpcErrors, rpcDurations)

	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			start, code := time.Now(), ""
			r := h(ctx, method, params)

			// log metrics
			if n := zenrpc.NamespaceFromContext(ctx); n != "" {
				method = n + "." + method
			}

			// set platform & version
			platform, version := PlatformFromContext(ctx), VersionFromContext(ctx)

			if r.Error != nil {
				code = strconv.Itoa(r.Error.Code)
				rpcErrors.WithLabelValues(method, code, platform, version).Inc()
			}

			rpcDurations.WithLabelValues(method, code, platform, version).Observe(time.Since(start).Seconds())

			return r
		}
	}
}
