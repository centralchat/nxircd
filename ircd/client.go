package ircd

// The reasoning behind this was so the web package could import and the services
// package can import without
import (
	"fmt"
	"time"

	"nxircd/interfaces"
)

const (
	capStateStart = iota
	capStateNeg   = iota
	capStateEnd   = iota
)

// We shoud do imbeding - but it breaks test
type IRCClient struct {
	Nick  string
	Ident string
	Name  string
	Host  string
}

// Client Holds an IRC Client
type Client struct {
	Messageable

	Nick  string
	Ident string
	Name  string
	Host  string
	Vhost string

	sock interfaces.Socket

	CTime time.Time
	ATime time.Time

	IP       string
	RealHost string
	// HostMask string

	registered bool

	Server *Server

	local bool

	capState   int
	capVersion int

	Channels *ChanList

	Modes *ModeList

	connected bool
}

// NewClient Returns a new IRC Client
func NewClient(server *Server, sock interfaces.Socket) *Client {
	ip := sock.IP()

	return &Client{
		IP:        ip,
		sock:      sock,
		Server:    server,
		local:     server.me,
		CTime:     time.Now(),
		ATime:     time.Now(),
		Modes:     NewModeList(),
		Channels:  NewChanList(),
		connected: true,
	}
}

// Run our client loop
func (c *Client) Run() {
	var err error

	clientPreflight(c)

	for err == nil {
		line, err := c.sock.Read()
		if err != nil {
			if c.connected {
				c.Quit(fmt.Sprintf("%s", err))
			}
			break
		}

		msg, _ := NewMessage(line)

		// We got a blank line
		// Its ok to continue
		if msg.Blank {
			continue
		}

		if cmd, found := clientCmdMap[msg.Command]; found {
			cmd.Run(c, msg)
		}
	}
}

func (c *Client) SetNick(nick string) {
	onick := c.Nick
	c.Nick = nick

	c.Server.Clients.Move(string(onick), c)
}

func (c *Client) ChangeNick(nick string) {
	onick := c.Nick
	c.Nick = nick

	// TODO: This is to deap its driving me crazy.
	c.Server.Clients.Move(onick, c)
}

func (c *Client) Prefix() string {
	return ""
}

func (c *Client) RealHostMask() string {
	return fmt.Sprintf("%s!%s@%s", c.Nick, c.Ident, c.RealHost)
}

func (c *Client) HostMask() string {
	return fmt.Sprintf("%s!%s@%s", c.Nick, c.Ident, c.Host)
}

func (c *Client) IPMask() string {
	return fmt.Sprintf("%s!%s@%s", c.Nick, c.Ident, c.IP)
}

func (c *Client) ApplyModes(ms ...Mode) {
	for _, m := range ms {
		c.Modes.Add(m, "")
	}
}

func (c *Client) Quit(msg string) {
	c.sock.Close()
	for _, channel := range c.Channels.list {
		channel.Quit(c, msg)
		c.Channels.Delete(channel)
	}
	c.Server.RemoveClient(c)
}

func (c *Client) Part(channel *Channel, msg string) {
	c.Channels.Delete(channel)
	channel.Part(c, msg)
}

func (c *Client) Send(prefix, cmd string, args ...string) error {
	m := MakeMessage(prefix, cmd, args...)
	return c.sendMessage(m)
}

func (c *Client) Reply(cmd string, args ...string) error {
	m := MakeMessage(c.Nick, cmd, args...)
	return c.sendMessage(m)
}

func (c *Client) PrivMsg(cli *Client, msg string) {
	c.Send(cli.Nick, "PRIVMSG", c.Nick, msg+" ")
}

func (c *Client) Notice(cli *Client, msg string) {
	c.Send(cli.Nick, "NOTICE", c.Nick, msg+" ")
}

func (c *Client) SendFromServer(cmd string, args ...string) error {
	m := MakeMessage(c.Server.Name, cmd, args...)
	return c.sendMessage(m)
}

func (c *Client) SendNumeric(num string, args ...string) error {
	a := []string{c.Nick}
	a = append(a, args...)

	m := MakeMessage(c.Server.Name, num, a...)
	return c.sendMessage(m)
}

// sendMessage -
func (c *Client) sendMessage(msg *Message) error {
	_, err := c.sock.Write(msg.String())
	return err
}
