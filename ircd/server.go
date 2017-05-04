package ircd

import (
	"fmt"
	"net"
	"nxircd/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	"nxircd/config"
)

var (
	// ServerSignals - The signals we respond to
	ServerSignals = []os.Signal{syscall.SIGINT, syscall.SIGHUP,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1}
)

// Listener short hand type for us
type Listener map[string]net.Listener

// Server is the main server struct for the IRCd
type Server struct {
	name    string
	network string

	ctime time.Time
	time  time.Time

	config *config.Config

	clients  *ClientList
	channels *ChannelList

	connections chan net.Conn
	signals     chan os.Signal

	listeners Listener
	log       *log.Logger
}

/**************************************************************/

// NewServer create an instance of server
func NewServer(config *config.Config) *Server {
	server := &Server{
		config:      config,
		name:        config.Name,
		network:     config.Network,
		ctime:       time.Now(),
		clients:     NewClientList(),
		channels:    NewChannelList(),
		connections: make(chan net.Conn),
		signals:     make(chan os.Signal, len(ServerSignals)),
		listeners:   make(Listener, 15),
		log:         log.New("ircd ", "nxircd.log", config.LogLevel),
	}
	signal.Notify(server.signals, ServerSignals...)

	server.bindListeners()

	return server
}

/**************************************************************/

// Run is the main loop of the application
func (server *Server) Run() {
	done := false
	for !done {
		server.time = time.Now()

		select {
		case signal := <-server.signals:
			if signal != syscall.SIGUSR1 {
				done = true
				continue
			}

			server.handleSignal(signal)
		case conn := <-server.connections:
			go NewClient(server, conn)
		}
	}
}

/**************************************************************/

func (server *Server) handleSignal(signal os.Signal) {
	switch {
	case signal == syscall.SIGUSR1:
		server.channels.lock.Lock()
		for _, channel := range server.channels.list {
			channel.lock.Lock()

			server.log.Info("Channel: %s", channel.name)
			server.log.Info("  Clients:")

			for _, cclient := range channel.clients {
				client := cclient.client
				server.log.Info("    %s [%s]", client.nickMask, client.ip)
			}
			channel.lock.Unlock()
		}
		server.channels.lock.Unlock()
	}
}

/**************************************************************/

func (server *Server) bindListeners() {
	for _, addr := range server.config.ListenersFor("ircd") {
		if addr.Type == "ircd" {
			listener, err := server.listen(addr.Host)
			if err != nil {
				fmt.Printf("Unable to create listener: %s\n", err)
				continue
			}

			server.listeners[addr.Host] = listener

			server.log.Info("Listening (%s): %s\n", addr.Type, addr.Host)
		}
	}
}

/**************************************************************/

func (server *Server) listen(addr string) (listener net.Listener, err error) {

	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				server.log.Warn("%s accept error: %s\n", server.name, err)
				continue
			}
			server.log.Debug("%s accept: %s\n", server.name, conn.RemoteAddr())

			server.connections <- conn
		}
	}()

	return
}

/**************************************************************/

func (server *Server) register(client *Client) {
	if client.state == clientStateCapNeg {
		return
	}
	client.Register()

	client.Send(server.name, RPL_WELCOME, client.nick, fmt.Sprintf("Welcome to the Internet Relay Network %s", client.nick))
	client.Send(server.name, RPL_YOURHOST, client.nick, fmt.Sprintf("Your host is %s, running version %s", server.name, VER_STRING))
	client.Send(server.name, RPL_CREATED, client.nick, fmt.Sprintf("This server was created %s", server.ctime.Format(time.RFC1123)))
	client.Send(server.name, RPL_MYINFO, client.nick, server.name, VER_STRING)
	client.Send(server.name, RPL_ISUPPORT, client.nick)

	server.log.Info("Client registered [%s]", client.nick)
}

/**************************************************************/
