package vt

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apisrv/pkg/db"
	"apisrv/pkg/embedlog"

	"github.com/vmkteam/zenrpc/v2"
)

type userCtx string

const (
	userKey userCtx = "vt.user"
)

func authMiddleware(commonRepo *db.CommonRepo, logger embedlog.Logger) zenrpc.MiddlewareFunc {
	return func(h zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			req, ok := zenrpc.RequestFromContext(ctx)
			if !ok {
				return h(ctx, method, params)
			}

			ns := zenrpc.NamespaceFromContext(ctx)

			// skip auth.Login method
			if ns == NSAuth && method == RPC.AuthService.Login {
				return h(ctx, method, params)
			}

			authHeader := req.Header.Get(AuthKey)
			// return error if header is not set
			if authHeader == "" {
				return zenrpc.NewResponseError(zenrpc.IDFromContext(ctx), ErrUnauthorized.Code, ErrUnauthorized.Message, ErrUnauthorized.Data)
			}

			// return error  if user not found
			dbu, err := commonRepo.EnabledUserByAuthKey(ctx, authHeader)
			if err != nil || dbu == nil {
				return zenrpc.NewResponseError(zenrpc.IDFromContext(ctx), ErrUnauthorized.Code, ErrUnauthorized.Message, ErrUnauthorized.Data)
			}

			// updating last activity
			if dbu.LastActivityAt == nil || time.Since(*dbu.LastActivityAt) > time.Second*90 {
				if _, err := commonRepo.UpdateUserActivity(ctx, dbu); err != nil {
					logger.Errorf("update user activity error=%s", err)
				}
			}

			return h(context.WithValue(ctx, userKey, dbu), method, params)
		}
	}
}

func UserFromContext(ctx context.Context) *db.User {
	if user, ok := ctx.Value(userKey).(*db.User); ok {
		return user
	}
	return nil
}

// HTTPAuthMiddleware checks user from authKey header
func HTTPAuthMiddleware(commonRepo db.CommonRepo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errCode := http.StatusUnauthorized

		// return error if header is not set
		authHeader := r.Header.Get(AuthKey)
		if authHeader == "" {
			http.Error(w, "authorization required", errCode)
			return
		}

		// return error if user not found
		dbu, err := commonRepo.EnabledUserByAuthKey(r.Context(), authHeader)
		if err != nil || dbu == nil {
			http.Error(w, "user not found", errCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}
