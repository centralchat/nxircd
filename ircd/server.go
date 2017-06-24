package ircd

import (
	"net"
	"time"

	"nxircd/config"

	"os"

	"github.com/apsdehal/go-logger"
)

// Server is the main server struct for the IRCd
type Server struct {
	Name    string
	Network string

	CTime time.Time
	Time  time.Time

	Log    *logger.Logger
	Config *config.Config

	connections chan net.Conn

	Signals chan os.Signal

	stopping bool
}

func NewServer(conf *config.Config, log *logger.Logger) *Server {
	return &Server{
		Name:        conf.Name,
		Network:     conf.Network,
		CTime:       time.Now(),
		Time:        time.Now(),
		connections: make(chan net.Conn),
		Signals:     make(chan os.Signal),
		Log:         log,
		Config:      conf,
	}
}

// Run our server
func (serv *Server) Run() {
	serv.bindListeners()
	for !serv.stopping {
		select {
		case conn := <-serv.connections:
			serv.Log.DebugF("Accepted connection: ", conn.RemoteAddr)
		case sig := <-serv.Signals:
			serv.stopping = true
			serv.Log.InfoF("Recieved: %v", sig)
		}
	}
}

func (serv *Server) bindListeners() {
	listeners := serv.Config.ListenersFor("ircd")
	for _, listener := range listeners {
		serv.listen(listener)
	}

}

func (serv *Server) listen(listen config.Listen) (net.Listener, error) {
	listener, err := net.Listen("tcp", listen.Host)
	if err != nil {
		return listener, err
	}

	serv.Log.InfoF("Listening on:  %s", listen.Host)

	go serv.acceptLoop(listener)
	return listener, nil
}

func (serv *Server) acceptLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		serv.connections <- conn
	}
}
