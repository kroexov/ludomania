package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vmkteam/zenrpc/v2"
)

type noCancel struct {
	ctx context.Context
}

func (c noCancel) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (c noCancel) Done() <-chan struct{}             { return nil }
func (c noCancel) Err() error                        { return nil }
func (c noCancel) Value(key interface{}) interface{} { return c.ctx.Value(key) }

// withoutCancel returns a context that is never canceled.
func withoutCancel(ctx context.Context) context.Context {
	return noCancel{ctx: ctx}
}

// WithNoCancelContext ignores Cancel func from context. This is useful for passing context to `go-pg`.
func WithNoCancelContext() zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			return h(withoutCancel(ctx), method, params)
		}
	}
}
