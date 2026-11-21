package cli

import (
	"os"
	"os/signal"
	"syscall"
)

func notifyLog(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

func stopNotifyLog(ch chan os.Signal) {
	signal.Stop(ch)
}
