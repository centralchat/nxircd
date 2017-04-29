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
// "nxircd/client"
)

var (
  // ServerSignals - The signals we respond to
  ServerSignals = []os.Signal{syscall.SIGINT, syscall.SIGHUP,
    syscall.SIGTERM, syscall.SIGQUIT}
)

// Listener short hand type for us
type Listener map[string]net.Listener

// Server is the main server struct for the IRCd
type Server struct {
  name string

  ctime  time.Time
  config *Config

  clients *ClientList

  connections chan net.Conn
  signals     chan os.Signal

  listeners Listener
  log       *log.Logger
}

/**************************************************************/

// NewServer create an instance of server
func NewServer(config *Config) *Server {
  server := &Server{
    ctime:       time.Now(),
    config:      config,
    clients:     NewClientList(),
    connections: make(chan net.Conn),
    signals:     make(chan os.Signal, len(ServerSignals)),
    listeners:   make(Listener, 15),
    log:         log.New("ircd ", "nxircd.log", config.LogLevel),
  }

  server.name = config.Name

  signal.Notify(server.signals, ServerSignals...)

  server.bindListeners()

  return server
}

/**************************************************************/

// Run is the main loop of the application
func (server *Server) Run() {
  done := false
  for !done {
    select {
    case <-server.signals:
      fmt.Printf("Recieved Signal?")
      // server.Shutdown()
      done = true

    case conn := <-server.connections:
      go NewClient(server, conn)
    }
  }
}

/**************************************************************/

func (server *Server) bindListeners() {
  for _, addr := range server.config.listenersFor("ircd") {
    if addr.Type == "ircd" {
      listener, err := server.listen(addr.Host)
      if err != nil {
        fmt.Printf("Unable to create listener: %s\n", err)
        continue
      }

      server.listeners[addr.Host] = listener

      fmt.Printf("Listening (%s): %s\n", addr.Type, addr.Host)
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

func (server *Server) register(client *Client) {
  if client.state == clientStateCapNeg {
    return
  }
  client.Register()

  client.Send(server.name, RPL_WELCOME, client.nick, fmt.Sprintf("Welcome to the Internet Relay Network %s", client.nick))
  client.Send(server.name, RPL_YOURHOST, client.nick, fmt.Sprintf("Your host is %s, running version %s", server.name, VER_STRING))
  client.Send(server.name, RPL_CREATED, client.nick, fmt.Sprintf("This server was created %s", server.ctime.Format(time.RFC1123)))
  client.Send(server.name, RPL_MYINFO, client.nick, server.name, VER_STRING, "", "")

  server.log.Info("Client registered [%s]", client.nick)
}

/**************************************************************/
