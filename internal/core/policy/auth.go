package policy

import (
	"bufio"
	"net"
	"net/http"
	"reflect"
	"strings"

	goauth "github.com/MrEthical07/goAuth"
	goauthmiddleware "github.com/MrEthical07/goAuth/middleware"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/params"
	"github.com/MrEthical07/superapi/internal/core/projectscope"
	"github.com/MrEthical07/superapi/internal/core/requestid"
	"github.com/MrEthical07/superapi/internal/core/response"
)

// AuthRequired enforces authentication on a route.
//
// It validates bearer token claims using the configured goAuth engine and
// injects auth.AuthContext into request context for downstream policies.
//
// Behavior:
// - Returns 401 if token is missing or invalid
// - Uses selected mode (jwt_only, hybrid, strict)
//
// Usage:
//
//	r.Handle(http.MethodGet, "/api/v1/system/whoami", httpx.Adapter(handler),
//	    policy.AuthRequired(engine, mode),
//	)
//
// Notes:
// - Place before RBAC and project policies
// - Requires a non-nil engine when route is expected to be protected
func AuthRequired(engine *goauth.Engine, mode auth.Mode) Policy {
	p := authRequiredWithEngine(engine, mode)
	return annotatePolicy(p, Metadata{Type: PolicyTypeAuthRequired, Name: "AuthRequired"})
}

type authGuardResponseWriter struct {
	http.ResponseWriter
	unauthorized         bool
	suppressUnauthorized func() bool
}

// WriteHeader suppresses premature 401 writes when guard short-circuits.
func (w *authGuardResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusUnauthorized && w.shouldSuppressUnauthorized() {
		w.unauthorized = true
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write discards body bytes for suppressed unauthorized responses.
func (w *authGuardResponseWriter) Write(data []byte) (int, error) {
	if w.unauthorized && w.shouldSuppressUnauthorized() {
		return len(data), nil
	}
	return w.ResponseWriter.Write(data)
}

func (w *authGuardResponseWriter) shouldSuppressUnauthorized() bool {
	if w == nil || w.suppressUnauthorized == nil {
		return true
	}
	return w.suppressUnauthorized()
}

// SetRoutePattern forwards route-pattern propagation when supported.
func (w *authGuardResponseWriter) SetRoutePattern(pattern string) {
	if setter, ok := w.ResponseWriter.(interface{ SetRoutePattern(string) }); ok {
		setter.SetRoutePattern(pattern)
	}
}

// Flush forwards flush capability when supported by underlying writer.
func (w *authGuardResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack forwards connection hijack when supported.
func (w *authGuardResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// Push forwards HTTP/2 server push when supported.
func (w *authGuardResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func authRequiredWithEngine(engine *goauth.Engine, mode auth.Mode) Policy {
	routeMode := goauth.ModeInherit
	switch mode {
	case auth.ModeJWTOnly:
		routeMode = goauth.ModeJWTOnly
	case auth.ModeStrict:
		routeMode = goauth.ModeStrict
	}

	guard := goauthmiddleware.Guard(engine, routeMode)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled := false
			guarded := guard(http.HandlerFunc(func(innerW http.ResponseWriter, innerR *http.Request) {
				nextCalled = true
				result, ok := goauthmiddleware.AuthResultFromContext(innerR.Context())
				if !ok || result == nil || strings.TrimSpace(result.UserID) == "" {
					rid := requestid.FromContext(innerR.Context())
					response.Error(innerW, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), rid)
					return
				}

				principal := auth.AuthContext{
					UserID:         result.UserID,
					Role:           result.Role,
					PermissionMask: maskValueToUint64(result.Mask),
					Permissions:    append([]string(nil), result.Permissions...),
				}

				ctx := auth.WithContext(innerR.Context(), principal)
				next.ServeHTTP(innerW, innerR.WithContext(ctx))
			}))

			adapter := &authGuardResponseWriter{
				ResponseWriter: w,
				suppressUnauthorized: func() bool {
					return !nextCalled
				},
			}
			guarded.ServeHTTP(adapter, r)
			if adapter.unauthorized && !nextCalled {
				rid := requestid.FromContext(r.Context())
				response.Error(w, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), rid)
			}
		})
	}
}

func maskValueToUint64(mask any) uint64 {
	v := reflect.ValueOf(mask)
	if !v.IsValid() {
		return 0
	}

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := v.Int()
		if i < 0 {
			return 0
		}
		return uint64(i)
	default:
		return 0
	}
}

// ProjectRequired ensures authenticated requests carry project scope.
//
// Behavior:
// - Returns 401 when authentication context is absent
// - Resolves project scope strictly from request path params (project_id/projectId)
// - Returns 403 when scope is missing or conflicting
//
// Notes:
// - Required for project-isolated routes
// - Place after AuthRequired
func ProjectRequired() Policy {
	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())
			principal, ok := auth.FromContext(r.Context())
			if !ok || strings.TrimSpace(principal.UserID) == "" {
				response.Error(w, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), rid)
				return
			}

			pathProjectID, hasPathProjectID := requestProjectID(r)
			if !hasPathProjectID {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope required"), rid)
				return
			}

			principalProjectID := strings.TrimSpace(principal.ProjectID)
			if principalProjectID != "" && !projectscope.IsSameProject(principalProjectID, pathProjectID) {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope mismatch"), rid)
				return
			}

			principal.ProjectID = pathProjectID

			next.ServeHTTP(w, r.WithContext(auth.WithContext(r.Context(), principal)))
		})
	}

	return annotatePolicy(p, Metadata{Type: PolicyTypeProjectRequired, Name: "ProjectRequired"})
}

// ProjectMatchFromPath enforces project isolation using a route path parameter.
//
// Behavior:
// - Returns 400 when route project parameter is missing
// - Returns 401 when auth context is missing
// - Returns 403 when principal project and route project mismatch
//
// Usage:
//
//	r.Handle(http.MethodGet, "/api/v1/projects/{project_id}/tasks", handler,
//	    policy.AuthRequired(engine, mode),
//	    policy.ProjectRequired(),
//	    policy.ProjectMatchFromPath("project_id"),
//	)
func ProjectMatchFromPath(paramName string) Policy {
	paramName = strings.TrimSpace(paramName)
	if paramName == "" {
		panicInvalidRouteConfigf("%s requires a non-empty path parameter name", PolicyTypeProjectMatchFromPath)
	}

	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())
			principal, ok := auth.FromContext(r.Context())
			if !ok {
				response.Error(w, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), rid)
				return
			}

			resourceProject := strings.TrimSpace(params.URLParam(r, paramName))
			if resourceProject == "" {
				response.Error(w, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, paramName+" is required"), rid)
				return
			}
			if strings.TrimSpace(principal.ProjectID) == "" {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope required"), rid)
				return
			}
			if !projectscope.IsSameProject(principal.ProjectID, resourceProject) {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope mismatch"), rid)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	return annotatePolicy(p, Metadata{
		Type:             PolicyTypeProjectMatchFromPath,
		Name:             "ProjectMatchFromPath",
		ProjectPathParam: paramName,
	})
}

func requestProjectID(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}

	candidates := []string{"project_id", "projectId"}
	for _, name := range candidates {
		if value := strings.TrimSpace(params.URLParam(r, name)); value != "" {
			return value, true
		}
		if value := strings.TrimSpace(r.PathValue(name)); value != "" {
			return value, true
		}
	}

	return "", false
}
