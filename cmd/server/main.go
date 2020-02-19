package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/yyklll/skeleton/pkg/api"
	"github.com/yyklll/skeleton/pkg/config"
	"github.com/yyklll/skeleton/pkg/log"
	"github.com/yyklll/skeleton/pkg/signals"
	"github.com/yyklll/skeleton/pkg/version"
)

func main() {
	// flags definition
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	fs.Int("port", 6666, "HTTP port")
	fs.Int("port-metrics", 8989, "metrics port")
	// fs.Int("grpc-port", 0, "gRPC port")
	// fs.String("grpc-service-name", "skeleton", "gPRC service name")
	fs.String("log-level", "info", "log level debug, info, warning, error or flatal")
	fs.Duration("http-client-timeout", 2*time.Minute, "client timeout duration")
	fs.Duration("http-server-timeout", 30*time.Second, "server read and write timeout duration")
	fs.Duration("http-server-shutdown-timeout", 5*time.Second, "server graceful shutdown timeout duration")
	fs.String("config-path", "", "config dir path")
	fs.String("config", "config.yaml", "config file name")
	// parse flags
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		fs.PrintDefaults()
		os.Exit(2)
	}

	// load configuration file
	config.SetupConfig("skeleton", fs)
	config.LoadConfig()

	// configure logging
	log.InitGlobalLogger(config.GetConfigField("log-level").(string))

	// load gRPC server config
	// var grpcCfg grpc.Config
	// if err := viper.Unmarshal(&grpcCfg); err != nil {
	// 	log.Fatal("config unmarshal failed", err)
	// }

	// // start gRPC server
	// if grpcCfg.Port > 0 {
	// 	grpcSrv, _ := grpc.NewServer(&grpcCfg, logger)
	// 	go grpcSrv.ListenAndServe()
	// }

	// load HTTP server config
	var cfg config.AllConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("config unmarshal failed", err)
	}

	// log version and port
	log.Info("Starting skeleton", "version", version.VERSION,
		"revision", version.REVISION,
		"port", cfg.SrvConfig.Port,
	)

	// start HTTP server
	srv := api.NewServer(&cfg)
	stopCh := signals.SetupSignalHandler()
	srv.ListenAndServe(stopCh)
}
