package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vmkteam/zenrpc/v2"
)

// WithAPILogger logs via Printf function (e.g. log.Printf) all requests.
func WithAPILogger(pf Printf, serverName string) zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			start := time.Now()
			r := h(ctx, method, params)

			methodName := fullMethodName(serverName, zenrpc.NamespaceFromContext(ctx), method)
			pf("ip=%s platform=%q version=%q method=%s duration=%v params=%q err=%q userAgent=%q xRequestId=%q",
				IPFromContext(ctx),
				PlatformFromContext(ctx),
				VersionFromContext(ctx),
				methodName,
				time.Since(start),
				params,
				r.Error,
				UserAgentFromContext(ctx),
				XRequestIDFromContext(ctx),
			)

			return r
		}
	}
}
