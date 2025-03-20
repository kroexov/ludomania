# zenrpc-middleware: Middlewares for vmkteam/zenrpc

[![Go Report Card](https://goreportcard.com/badge/github.com/vmkteam/zenrpc-middleware)](https://goreportcard.com/report/github.com/vmkteam/zenrpc-middleware) [![Go Reference](https://pkg.go.dev/badge/github.com/vmkteam/zenrpc-middleware.svg)](https://pkg.go.dev/github.com/vmkteam/zenrpc-middleware)

`zenrpc-middleware` is a set of common middlewares for [zenrpc](https://github.com/vmkteam/zenrpc) implementing logging,
metrics and error tracking.

## Middlewares

### WithDevel

Sets bool flag to context for detecting development environment.

### WithHeaders
    
Sets UserAgent, Platform, Version to context. UserAgent strips to 2048 chars, Platform and Version – to 64.

### WithAPILogger

Logs via Printf function (e.g. log.Printf) all requests. For example

```text
ip= platform="" version="" method=nodevel.arith.divide duration=63.659µs params="{ \"a\": 1, \"b\": 24 }" err=<nil> userAgent="Go-http-client/1.1"
```

### WithSentry

Sets additional parameters for current Sentry scope. Extras: params, duration, ip. Tags: platform,
version, method.

### WithNoCancelContext

Ignores Cancel func from context. This is useful for passing context to `go-pg`.

### WithMetrics

Logs duration of RPC requests via Prometheus. Default AppName is zenrpc. It exposes two
metrics: `appName_rpc_error_requests_count` and `appName_rpc_responses_duration_seconds`. Labels: method, code,
platform, version.

### WithTiming

Adds timings in JSON-RPC 2.0 Response via `extensions` field (not in spec). Middleware is active
when `isDevel=true` or AllowDebugFunc returns `true`. Sample AllowDebugFunc (checks GET/POST parameters for "true"
value, like `?d=true`):

```go
allowDebugFn := func (param string) middleware.AllowDebugFunc {
    return func (req *http.Request) bool {
        return req.FormValue(param) == "true"
    }
}
```

`DurationLocal` – total method execution time in ms.

If `DurationRemote` or `DurationDiff` are set then `DurationLocal` excludes these values.

### WithSQLLogger

Adds `SQL` or `DurationSQL` fields in JSON-RPC 2.0 Response `extensions` field (not in spec).

`DurationSQL` field is set then  `isDevel=true` or AllowDebugFunc(allowDebugFunc) returns `true`.

`SQL` field is set then `isDevel=true` or AllowDebugFunc(allowDebugFunc, allowSqlDebugFunc) returns `true`.

### WithErrorLogger

Logs all errors (ErrorCode==500 or < 0) via Printf func and sends them to Sentry. It also removes
sensitive error data from response. It is good to use pkg/errors for stack trace support in sentry.

```text
ip= platform="Test1" version="v1.0.0 alpha1" method=errordevel.arith.checkzenrpcerror duration=30.786µs params=[ true ] err="test"
```

### XRequestID

Handler for adding `X-Request-ID` header to all requests and responses.

```go
	s := http.NewServer(middleware.XRequestID(http.HandlerFunc(rpc.ServeHTTP)))
```

### EchoIPContext

Echo middleware for adding client IP info to context for zenrpc middleware

```go
    e.Use(middleware.EchoIPContext())
```

### EchoSentryHubContext

Echo middleware for adding sentry hub info to context for zenrpc middleware

```go
    e.Use(middleware.EchoSentryHubContext())
```

## Examples

### Basic usage

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/vmkteam/zenrpc-middleware"
	"github.com/vmkteam/zenrpc/v2"
)

func main() {
	dbс := pg.Connect(&pg.Options{User: "postgres"})
	defer dbс.Close()

	isDevel := true
	elog := log.New(os.Stderr, "E", log.LstdFlags|log.Lshortfile)
	dlog := log.New(os.Stdout, "D", log.LstdFlags|log.Lshortfile)

	allowDebug := func(param string) middleware.AllowDebugFunc {
		return func(req *http.Request) bool {
			return req.FormValue(param) == "true"
		}
	}

	rpc := zenrpc.NewServer(zenrpc.Options{
		ExposeSMD: true,
		AllowCORS: true,
	})

	rpc.Use(
		middleware.WithDevel(isDevel),
		middleware.WithHeaders(),
		middleware.WithAPILogger(dlog.Printf, middleware.DefaultServerName),
		middleware.WithSentry(middleware.DefaultServerName),
		middleware.WithNoCancelContext(),
		middleware.WithMetrics(middleware.DefaultServerName),
		middleware.WithTiming(isDevel, allowDebug("d")),
		middleware.WithSQLLogger(dbc, isDevel, allowDebug("d"), allowDebug("s")),
		middleware.WithErrorLogger(elog.Printf, middleware.DefaultServerName),
	)
}


    // rpc.Register and server
```

### Handler snippets

```go
// sentry init
sentry.Init(sentry.ClientOptions{
    Dsn:         cfg.Sentry.DSN,
    Environment: cfg.Sentry.Environment,
    Release:     version,
}

// sentry middleware for Echo
e.Use(sentryecho.New(sentryecho.Options{
    Repanic:         true,
    WaitForDelivery: true,
}))

// register handler
a.echo.Any("/v1/rpc/", middleware.EchoHandler(rpc))

---  OR ---

e.Use(middleware.EchoIPContext()), middleware.EchoSentryHubContext())
// register handler
e.Any("/int/rpc/", echo.WrapHandler(XRequestID(rpc)))

```
