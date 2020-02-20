package api

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	_ "github.com/yyklll/skeleton/docs"
	"github.com/yyklll/skeleton/pkg/auth"
	"github.com/yyklll/skeleton/pkg/config"
	"github.com/yyklll/skeleton/pkg/log"
	"github.com/yyklll/skeleton/pkg/watcher"
)

// @title Skeleton API Service
// @version 0.1
// @description Go microservice template for Kubernetes.

// @host localhost:6666
// @BasePath /
// @schemes http https

var (
	healthy int32
	ready   int32
	// watcher *watcher.Watcher
)

type Server struct {
	router *mux.Router
	config *config.AllConfig
}

func NewServer(config *config.AllConfig) *Server {
	return &Server{
		router: mux.NewRouter(),
		config: config,
	}
}

func (s *Server) registerHandlers(a auth.AuthenticationDriver) {
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	s.router.HandleFunc("/info", s.infoHandler).Methods("GET")
	s.router.HandleFunc("/version", s.versionHandler).Methods("GET")
	s.router.HandleFunc("/healthz", s.healthzHandler).Methods("GET")
	s.router.HandleFunc("/readyz", s.readyzHandler).Methods("GET")
	s.router.HandleFunc("/readyz/enable", s.enableReadyHandler).Methods("POST")
	s.router.HandleFunc("/readyz/disable", s.disableReadyHandler).Methods("POST")
	s.router.HandleFunc("/token", a.TokenGenerateHandler).Methods("POST")
}

func (s *Server) registerMiddlewares(a auth.AuthenticationDriver) {
	prom := NewPrometheusMiddleware()
	s.router.Use(prom.Handler)
	s.router.Use(s.reqLoggingMiddleware)
	s.router.Use(a.TokenValidateMiddleware)
}

func (s *Server) ListenAndServe(stopCh <-chan struct{}) {
	go s.startMetricsServer()

	authDriver, _ := auth.Create("jwt", map[string]interface{}{"secret": s.config.SrvConfig.JWTSecret})
	s.registerHandlers(authDriver)
	s.registerMiddlewares(authDriver)

	var handler http.Handler
	if s.config.SrvConfig.H2C {
		handler = h2c.NewHandler(s.router, &http2.Server{})
	} else {
		handler = s.router
	}

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(s.config.SrvConfig.Port),
		WriteTimeout: s.config.SrvConfig.HttpServerTimeout,
		ReadTimeout:  s.config.SrvConfig.HttpServerTimeout,
		IdleTimeout:  2 * s.config.SrvConfig.HttpServerTimeout,
		Handler:      handler,
	}

	// s.printRoutes()

	// load configs in memory and start watching for changes in the config dir
	if stat, err := os.Stat(s.config.SrvConfig.ConfigPath); err == nil && stat.IsDir() {
		cwatcher, err := watcher.NewWatch(s.config.SrvConfig.ConfigPath)
		if err != nil {
			log.Error("config watch error", err, "config path", s.config.SrvConfig.ConfigPath)
		} else {
			cwatcher.Watch()
		}
	}

	// run server in background
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP server start error", err)
		}
	}()

	// signal Kubernetes the server is ready to receive traffic
	atomic.StoreInt32(&healthy, 1)
	atomic.StoreInt32(&ready, 1)

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), s.config.SrvConfig.HttpServerShutdownTimeout)
	defer cancel()

	// all calls to /readyz will fail from now on
	atomic.StoreInt32(&ready, 0)

	log.Info("Shutting down HTTP server, timeout", s.config.SrvConfig.HttpServerShutdownTimeout)

	// wait for Kubernetes readiness probe to remove this instance from the load balancer
	// the readiness check interval must be lower than the timeout
	if viper.GetString("log-level") != "debug" {
		time.Sleep(3 * time.Second)
	}

	// attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Warning("HTTP server graceful shutdown failed", err)
	} else {
		log.Info("HTTP server has already stopped")
	}
}

func (s *Server) startMetricsServer() {
	if s.config.SrvConfig.PortMetrics > 0 {
		mux := http.DefaultServeMux
		mux.Handle("/metrics", promhttp.Handler())

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%v", s.config.SrvConfig.PortMetrics),
			Handler: mux,
		}

		go srv.ListenAndServe()
	}
}

func (s *Server) printRoutes() {
	s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
}

type ArrayResponse []string
type MapResponse map[string]string
