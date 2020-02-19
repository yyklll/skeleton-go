package signals

import (
	"os"
	"os/signal"

	"github.com/yyklll/skeleton/pkg/log"
)

var onlyOneSignalHandler = make(chan struct{})

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		log.Info("Signal received, shutting down the server...")
		close(stop)
		<-c
		log.Info("A second signal received, shutting down the server immediately")
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}
