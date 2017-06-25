package main

import (
	"fmt"
	"os"
	"syscall"

	"nxircd/config"
	"nxircd/ircd"

	"os/signal"

	"github.com/apsdehal/go-logger"
)

var (
	// ServerSignals - The signals we respond to
	ServerSignals = []os.Signal{syscall.SIGINT, syscall.SIGHUP,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1}
)

type NxIRC struct {
	IRCServer *ircd.Server
	Config    *config.Config
	Log       *logger.Logger
	signals   chan os.Signal
}

func main() {
	logger.New("test", 1, os.Stdout)
	conf, err := config.New("config.json")
	if err != nil {
		fmt.Println("Unable to load in config: ", err)
		return
	}

	fmt.Println(conf)

	nxircd := NewNxIRC(conf)
	nxircd.Run()
}

// NewNxIRC - Create a new instance of NxIRC for consumption
func NewNxIRC(conf *config.Config) *NxIRC {
	logger, _ := logger.New("test", 1, os.Stdout)

	server := ircd.NewLocalServer(conf, logger, true)

	nx := &NxIRC{
		Config:    conf,
		IRCServer: server,
		Log:       logger,
		signals:   make(chan os.Signal),
	}

	signal.Notify(nx.signals, ServerSignals...)
	return nx
}

// Run - Runs our nxIRC service watching and dispatching signals
// to the various components
func (nx *NxIRC) Run() {
	nx.Banner()

	go nx.IRCServer.Run()
	done := false
	for !done {
		select {
		case sig := <-nx.signals:
			nx.IRCServer.Signals <- sig
			done = true
			continue
		}
	}

}

func (nx *NxIRC) Banner() {
	nx.Log.Notice("Booting nxIRCd")
}
