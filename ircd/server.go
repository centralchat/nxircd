package ircd

import (
	"net"
	"time"

	"nxircd/config"
	"nxircd/interfaces"

	"os"

	"github.com/apsdehal/go-logger"
)

// Server is the main server struct for the IRCd
type Server struct {
	Messageable

	Name    string
	Network string

	CTime time.Time
	Time  time.Time

	Log    *logger.Logger
	Config *config.Config

	connections chan interfaces.Socket

	Signals  chan os.Signal
	stopping bool

	Clients *ClientList

	ticker *time.Ticker

	me bool
}

// NewLocalServer - Create a ptr to a server struct
func NewLocalServer(conf *config.Config, log *logger.Logger, isMe bool) *Server {
	return &Server{
		Name:        conf.Name,
		Network:     conf.Network,
		CTime:       time.Now(),
		Time:        time.Now(),
		Log:         log,
		Config:      conf,
		Clients:     NewClientList(),
		connections: make(chan interfaces.Socket),
		Signals:     make(chan os.Signal),
		ticker:      time.NewTicker(1 * time.Second),
		me:          isMe,
	}
}

// Prefix - Returns the prefix as defined my Messageable interface
func (serv *Server) Prefix() string {
	return serv.Name
}

// Run our server
func (serv *Server) Run() {
	serv.bindListeners()
	for !serv.stopping {
		select {
		case sock := <-serv.connections:
			cli := NewClient(serv, sock)
			serv.Clients.Add(cli)
			go cli.Run()
		case sig := <-serv.Signals:
			serv.stopping = true
			serv.Log.InfoF("Recieved: %v", sig)
		case <-serv.ticker.C:
			// Adjust time
			serv.Time = time.Now()
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
		serv.connections <- NewIRCSocket(conn)
	}
}
