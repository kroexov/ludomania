package vt

import (
	"net/http"

	"apisrv/pkg/db"
	"apisrv/pkg/embedlog"

	zm "github.com/vmkteam/zenrpc-middleware"
	"github.com/vmkteam/zenrpc/v2"
)

//go:generate zenrpc

const (
	AuthKey = "Authorization2"
)

const (
	NSAuth = "auth"
	NSUser = "user"
)

var (
	ErrUnauthorized   = httpAsRpcError(http.StatusUnauthorized)
	ErrForbidden      = httpAsRpcError(http.StatusForbidden)
	ErrNotFound       = httpAsRpcError(http.StatusNotFound)
	ErrInternal       = httpAsRpcError(http.StatusInternalServerError)
	ErrNotImplemented = httpAsRpcError(http.StatusNotImplemented)
)

var allowDebugFn = func() zm.AllowDebugFunc {
	return func(req *http.Request) bool {
		return req != nil && req.FormValue("__level") == "5"
	}
}

func httpAsRpcError(code int) *zenrpc.Error {
	return zenrpc.NewStringError(code, http.StatusText(code))
}

// New returns new zenrpc Server.
func New(dbo db.DB, logger embedlog.Logger, isDevel bool) zenrpc.Server {
	rpc := zenrpc.NewServer(zenrpc.Options{
		ExposeSMD: true,
		AllowCORS: true,
	})

	commonRepo := db.NewCommonRepo(dbo)

	// middleware
	rpc.Use(
		authMiddleware(&commonRepo, logger),
		zm.WithDevel(isDevel),
		zm.WithHeaders(),
		zm.WithSentry(zm.DefaultServerName),
		zm.WithNoCancelContext(),
		zm.WithMetrics("vt"),
		zm.WithTiming(isDevel, allowDebugFn()),
		zm.WithSQLLogger(dbo.DB, isDevel, allowDebugFn(), allowDebugFn()),
	)

	if errlog, stdlog := logger.Loggers(); errlog != nil && stdlog != nil {
		rpc.Use(
			zm.WithAPILogger(stdlog.Printf, zm.DefaultServerName),
			zm.WithErrorLogger(errlog.Printf, zm.DefaultServerName),
		)
	}

	// services
	rpc.RegisterAll(map[string]zenrpc.Invoker{
		NSAuth: NewAuthService(dbo, logger),
		NSUser: NewUserService(dbo, logger),
	})

	return rpc
}
