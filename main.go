package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/raphael-p/brendan/config"
	configparser "github.com/raphael-p/gocommon/config"
	"github.com/raphael-p/gocommon/logger"
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

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Milliseconds()
		logger.Info(fmt.Sprintf(
			"%s %v from %v -> %d (%dms)",
			r.Method, r.URL, r.RemoteAddr, ww.Status(), duration,
		))
	})
}

// define app
func main() {
	// get working directory to be used to locate config and log files
	var workingDir string
	if os.Args[0] == "brendan" {
		ex, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf("failed to locate executable: %s", err))
		}
		workingDir = filepath.Dir(ex)
	} else {
		workingDir = "."
	}

	logger.Create(workingDir)
	defer logger.Close()
	configparser.Parse(workingDir, config.Envars.ConfigFilepath, config.Values, false)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logRequest)
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
			logger.Trace(fmt.Sprint("port: ", config.Values.Server.Port))
			w.Write([]byte("hello world"))
		})

	})

	http.ListenAndServe(fmt.Sprint(":", config.Values.Server.Port), r)
}
