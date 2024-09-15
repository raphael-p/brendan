package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/raphael-p/brendan/config"
	"github.com/raphael-p/gocommon"
)

// define app middleware
func allowLocal(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Milliseconds()
		gocommon.LogInfo(fmt.Sprintf(
			"%s %v from %v -> %d (%dms)",
			r.Method, r.URL, r.RemoteAddr, ww.Status(), duration,
		))
	})
}

// define app
func main() {
	workingDir := gocommon.GetExecDirectory("brendan")
	if workingDir == "" {
		workingDir = "."
	}
	gocommon.InitLogger(workingDir)
	defer gocommon.CloseLogger()
	gocommon.InitialiseConfig(workingDir, config.Envars.ConfigFilepath, config.Config)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logger)
	r.Use(middleware.Recoverer)

	// web endpoint
	r.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELLO WORLD!"))
	})

	// internal endpoints; access should be prevented by reverse proxy
	// server as well, this is just an extra layer of security
	r.Group(func(r chi.Router) {
		r.Use(allowLocal)

		r.Get("/private", func(w http.ResponseWriter, r *http.Request) {
			gocommon.LogTrace(fmt.Sprint("port: ", config.Config.Server.Port))
			w.Write([]byte("hello world"))
		})

	})

	http.ListenAndServe(fmt.Sprint(":", config.Config.Server.Port), r)
}
