package ircd

import (
	"fmt"
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
	messages    chan *Message
	Signals     chan os.Signal

	stopping bool

	Clients  *ClientList
	Channels *ChanList

	ticker *time.Ticker

	me bool
}

// NewLocalServer - Create a ptr to a server struct
func NewLocalServer(conf *config.Config, log *logger.Logger, isMe bool) *Server {
	return &Server{
		Name:    conf.Name,
		Network: conf.Network,

		CTime: time.Now(),
		Time:  time.Now(),

		Log:    log,
		Config: conf,

		Clients:  NewClientList(),
		Channels: NewChanList(),

		connections: make(chan interfaces.Socket),
		// Maybe we remove this?
		messages: make(chan *Message),
		Signals:  make(chan os.Signal),

		ticker: time.NewTicker(1 * time.Second),

		me: isMe,
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

func (serv *Server) NickInUse(nick string) bool {
	if cli := serv.Clients.Find(nick); cli != nil {
		return true
	}
	return false
}

func (serv *Server) Greet(cli *Client) {
	cli.SendNumeric(RPL_WELCOME, fmt.Sprintf("Welcome to the Internet Relay Network %s", cli.Nick))
	cli.SendNumeric(RPL_YOURHOST, fmt.Sprintf("Your host is %s, running version %s", serv.Name, VER_STRING))
	cli.SendNumeric(RPL_CREATED, fmt.Sprintf("This server was created %s", serv.CTime.Format(time.RFC1123)))
	cli.SendNumeric(RPL_MYINFO, VER_STRING)
	cli.SendNumeric(RPL_ISUPPORT, "")

	// cli.Send(server.name, RPL_WELCOME, client.nick, fmt.Sprintf("Welcome to the Internet Relay Network %s", client.nick))
	// cli.Send(server.name, RPL_YOURHOST, client.nick, fmt.Sprintf("Your host is %s, running version %s", server.name, VER_STRING))
	// cli.Send(server.name, RPL_CREATED, client.nick, fmt.Sprintf("This server was created %s", server.ctime.Format(time.RFC1123)))
	// cli.Send(server.name, RPL_MYINFO, client.nick, server.name, VER_STRING)
	// cli.Send(server.name, RPL_ISUPPORT, client.nick)
}
