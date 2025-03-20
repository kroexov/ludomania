package middleware

import (
	"context"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/vmkteam/zenrpc/v2"
)

const (
	isDevelCtx      contextKey = "isDevel"
	ctxPlatformKey  contextKey = "platform"
	ctxVersionKey   contextKey = "version"
	ctxIPKey        contextKey = "ip"
	ctxUserAgentKey contextKey = "userAgent"

	ctxNotificationKey = "JSONRPC2-Notification"

	DefaultServerName = ""
)

type (
	contextKey string
	Printf     func(format string, v ...interface{})
)

// WithDevel sets bool flag to context for detecting development environment.
func WithDevel(isDevel bool) zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			ctx = NewIsDevelContext(ctx, isDevel)
			return h(ctx, method, params)
		}
	}
}

// WithHeaders sets UserAgent, Platform, Version to context. UserAgent strips to 2048 chars, Platform and Version – to 64.
func WithHeaders() zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			if req, ok := zenrpc.RequestFromContext(ctx); ok && req != nil {
				ctx = NewUserAgentContext(ctx, req.UserAgent())
				ctx = NewPlatformContext(ctx, req.Header.Get("Platform"))
				ctx = NewVersionContext(ctx, req.Header.Get("Version"))
				ctx = NewXRequestIDContext(ctx, req.Header.Get(echo.HeaderXRequestID))
			}
			return h(ctx, method, params)
		}
	}
}

// NewIPContext creates new context with IP.
func NewIPContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ctxIPKey, ip)
}

// IPFromContext returns IP from context.
func IPFromContext(ctx context.Context) string {
	r, _ := ctx.Value(ctxIPKey).(string)
	return r
}

// NewUserAgentContext creates new context with User-Agent.
func NewUserAgentContext(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, ctxUserAgentKey, cutString(ua, 2048))
}

// UserAgentFromContext returns userAgent from context.
func UserAgentFromContext(ctx context.Context) string {
	r, _ := ctx.Value(ctxUserAgentKey).(string)
	return r
}

// NewNotificationContext creates new context with JSONRPC2 notification flag.
func NewNotificationContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxNotificationKey, true)
}

// NotificationFromContext returns JSONRPC2 notification flag from context.
func NotificationFromContext(ctx context.Context) bool {
	r, _ := ctx.Value(ctxNotificationKey).(bool)
	return r
}

// NewIsDevelContext creates new context with isDevel flag.
func NewIsDevelContext(ctx context.Context, isDevel bool) context.Context {
	return context.WithValue(ctx, isDevelCtx, isDevel)
}

// IsDevelFromContext returns isDevel flag from context.
func IsDevelFromContext(ctx context.Context) bool {
	if isDevel, ok := ctx.Value(isDevelCtx).(bool); ok {
		return isDevel
	}
	return false
}

// NewPlatformContext creates new context with platform.
func NewPlatformContext(ctx context.Context, platform string) context.Context {
	return context.WithValue(ctx, ctxPlatformKey, cutString(platform, 64))
}

// PlatformFromContext returns platform from context.
func PlatformFromContext(ctx context.Context) string {
	r, _ := ctx.Value(ctxPlatformKey).(string)
	return r
}

// NewVersionContext creates new context with version.
func NewVersionContext(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, ctxVersionKey, cutString(version, 64))
}

// VersionFromContext returns version from context.
func VersionFromContext(ctx context.Context) string {
	r, _ := ctx.Value(ctxVersionKey).(string)
	return r
}

// cutString cuts string with given length.
func cutString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

//fullMethodName returns namespace.method or serverName.namespace.method.
func fullMethodName(serverName, namespace, method string) string {
	name := namespace + "." + method
	if serverName != "" {
		name = serverName + "." + name
	}

	return name
}
