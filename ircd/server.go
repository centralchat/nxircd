package ircd

import (
	"fmt"
	"net"
	"time"

	"nxircd/config"
	"nxircd/interfaces"

	"os"

	"strconv"

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

	MaxClients int

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

		MaxClients: 0,

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
			serv.AddClient(cli)

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

		serv.Log.DebugF("Accepting Connection: %v", conn.RemoteAddr)
		serv.connections <- NewIRCSocket(conn)
	}
}

func (serv *Server) AddClient(cli *Client) {
	serv.Clients.Add(cli)

	cnt := serv.Clients.Count()
	if cnt > serv.MaxClients {
		serv.MaxClients = cnt
		serv.Log.InfoF("New max client count: %d", cnt)
	}
}

// FindClient - an alias to serv.Clients.Find for shorter code
func (serv *Server) FindClient(nick string) *Client {
	return serv.Clients.Find(nick)
}

// FindChannel - an alias to serv.Channels.Find for shorter code
func (serv *Server) FindChannel(name string) *Channel {
	return serv.Channels.Find(name)
}

// FindOrAddChan - Finds a channel , if that doesnt exist it adds it to the server registry
func (serv *Server) FindOrAddChan(name string) (ch *Channel) {
	ch = serv.FindChannel(name)
	if ch != nil {
		return
	}
	ch = NewChannel(name)
	serv.Channels.Add(ch)
	return
}

// RemoveClient - Remove a client from the server list
// Todo add additional logic etc here.
func (serv *Server) RemoveClient(cli *Client) {
	serv.Clients.Delete(cli)

	serv.Log.DebugF("New client count: %d", serv.Clients.Count())
}

func (serv *Server) NickInUse(nick string) bool {
	if cli := serv.Clients.Find(nick); cli != nil {
		return true
	}
	return false
}

func (serv *Server) LUsers(cli *Client) {
	cli.SendNumeric(RPL_LUSERCLIENT, fmt.Sprintf("There are %d users on 1 server", serv.Clients.Count()))
	cli.SendNumeric(RPL_LUSEROP, "0", "operator(s) online")
	cli.SendNumeric(RPL_LUSERCHANNELS, strconv.Itoa(serv.Channels.Count()), "channels formed")
	cli.SendNumeric(RPL_LUSERME, fmt.Sprintf("I have %d clients and 0 servers", serv.Clients.Count()))
}

func (serv *Server) Greet(cli *Client) {
	cli.SendNumeric(RPL_WELCOME, fmt.Sprintf("Welcome to the %s IRC Network %s", serv.Network, cli.RealHostMask()))
	cli.SendNumeric(RPL_YOURHOST, fmt.Sprintf("Your host is %s, running version %s", serv.Name, VERSION))
	cli.SendNumeric(RPL_CREATED, fmt.Sprintf("This server was created %s", serv.CTime.Format(time.RFC1123)))

	SendSupportLine(cli)

	cli.SendNumeric(RPL_USERHOST, fmt.Sprintf("%s is now your displayed host", cli.HostMask()))

	serv.LUsers(cli)
}

func NewTestClient(server *Server, nick, ident string) (*interfaces.TestSocket, *Client) {
	socket := interfaces.NewTestSocket()
	cli := NewClient(server, socket)

	cli.Nick = nick
	cli.Ident = ident
	cli.Host = cli.IP

	return socket, cli
}
