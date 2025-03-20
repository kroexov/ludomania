package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/labstack/echo/v4"
)

const ctxXRequestIDKey = echo.HeaderXRequestID

var xRequestIDre = regexp.MustCompile(`[a-zA-Z0-9-]+`)

func generateXRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func isValidXRequestID(requestId string) bool {
	return requestId != "" && len(requestId) <= 32 && xRequestIDre.Match([]byte(requestId))
}

// XRequestIDFromContext returns X-Request-ID from context.
func XRequestIDFromContext(ctx context.Context) string {
	r, _ := ctx.Value(ctxXRequestIDKey).(string)
	return r
}

// NewXRequestIDContext creates new context with X-Request-ID.
func NewXRequestIDContext(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, ctxXRequestIDKey, requestId)
}
