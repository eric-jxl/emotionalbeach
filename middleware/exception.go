package middleware

import (
	"log/slog"
	"net/http"
)

// RecoverMiddleware 在“边界层（接入层/异步入口/任务线程/流式循环）”强制 recover
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic_recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"trace_id", r.Header.Get("X-Trace-Id"))
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"code":500,"msg":"internal error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
