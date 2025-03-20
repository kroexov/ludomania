package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/vmkteam/zenrpc/v2"
)

const ctxSentryHubKey contextKey = "sentryHub"

// NewSentryHubContext creates new context with Sentry Hub.
func NewSentryHubContext(ctx context.Context, sentryHub *sentry.Hub) context.Context {
	if sentryHub == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxSentryHubKey, sentryHub)
}

// sentryHubFromContext returns Sentry Hub from context.
func sentryHubFromContext(ctx context.Context) (*sentry.Hub, bool) {
	r, ok := ctx.Value(ctxSentryHubKey).(*sentry.Hub)
	return r, ok
}

// WithSentry sets additional parameters for current Sentry scope. Extras: params, duration, ip. Tags: platform,
// version, method.
func WithSentry(serverName string) zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			if hub, ok := sentryHubFromContext(ctx); ok {
				start, platform, version, ip, xRequestID := time.Now(), PlatformFromContext(ctx), VersionFromContext(ctx), IPFromContext(ctx), XRequestIDFromContext(ctx)

				methodName := fullMethodName(serverName, zenrpc.NamespaceFromContext(ctx), method)

				hub.Scope().SetExtras(map[string]interface{}{
					"params":   params,
					"duration": time.Since(start).String(),
					"ip":       ip,
				})
				hub.Scope().SetTags(map[string]string{
					"platform":   platform,
					"version":    version,
					"method":     methodName,
					"xRequestId": xRequestID,
				})
			}

			return h(ctx, method, params)
		}
	}
}

// WithErrorLogger logs all errors (ErrorCode==500 or < 0) via Printf func and sends them to Sentry. It also removes
// sensitive error data from response. It is good to use pkg/errors for stack trace support in sentry.
func WithErrorLogger(pf Printf, serverName string) zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			start, platform, version, ip, xRequestID := time.Now(), PlatformFromContext(ctx), VersionFromContext(ctx), IPFromContext(ctx), XRequestIDFromContext(ctx)
			namespace := zenrpc.NamespaceFromContext(ctx)

			r := h(ctx, method, params)
			if r.Error != nil && (r.Error.Code == http.StatusInternalServerError || r.Error.Code < 0) {
				duration := time.Since(start)
				methodName := fullMethodName(serverName, namespace, method)

				pf("ip=%s platform=%q version=%q method=%s duration=%v params=%s xRequestId=%q err=%q", ip, platform, version, methodName, duration, params, xRequestID, r.Error)

				sentry.WithScope(func(scope *sentry.Scope) {
					scope.SetExtras(map[string]interface{}{
						"params":     params,
						"duration":   duration.String(),
						"ip":         ip,
						"error.data": r.Error.Data,
						"error.code": r.Error.Code,
					})
					scope.SetTags(map[string]string{
						"platform":   platform,
						"version":    version,
						"method":     methodName,
						"xRequestId": xRequestID,
					})
					sentry.CaptureException(r.Error)
				})

				// remove sensitive error data from response
				r.Error.Err = nil
				r.Error.Message = "Internal error"
			}

			return r
		}
	}
}
