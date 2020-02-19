package api

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/yyklll/skeleton/pkg/config"
)

func NewMockServer() *Server {
	config := &config.AllConfig{
		SrvConfig: config.ServerConfig{
			HttpClientTimeout:         30 * time.Second,
			HttpServerTimeout:         30 * time.Second,
			HttpServerShutdownTimeout: 5 * time.Second,
			ConfigPath:                "/config",
			Port:                      6666,
			PortMetrics:               9999,
			Hostname:                  "localhost",
			H2C:                       true,
			JWTSecret:                 "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			LogLevel:                  "debug",
		},
		Db: config.Database{
			Enable: false,
		},
	}

	return &Server{
		router: mux.NewRouter(),
		config: config,
	}
}
