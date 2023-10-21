package signal

import (
	"os"
	"os/signal"
	"syscall"
)

// ShutDownSingals returns all the singals that are being watched for to shut down services.
func ShutdownSignals() []os.Signal {
	return []os.Signal{
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
	}
}

func WaitShutdown() os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, ShutdownSignals()...)
	return <-signals
}
