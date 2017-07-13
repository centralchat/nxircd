package ircd

// The reasoning behind this was so the web package could import and the services
// package can import without
import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"time"

	"nxircd/interfaces"
	"strings"
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

// This is ugly but im tired so lets leave it be for now
func (c *Client) ApplyModeChanges(changes []ModeChange) {

	addString := ""
	delString := ""

	for _, change := range changes {
		switch change.Action {
		case ModeActionAdd:
			c.Modes.Add(change.Mode, "")
			addString += change.Mode.String()
		case ModeActionDel:
			c.Modes.Delete(change.Mode)
			delString += change.Mode.String()
		case ModeActionList:
			// Ignore for now
			continue
		}
	}

	if addString != "" {
		c.Send(c.Nick, "MODE", c.Nick, "+"+addString)
	}
	if delString != "" {
		c.Send(c.Nick, "MODE", c.Nick, "-"+delString)
	}
}

func (c *Client) RealHostMask() string {
	return fmt.Sprintf("%s!%s@%s", c.Nick, c.Ident, c.RealHost)
}

func (c *Client) SetMaskedHost() {
	if c.RealHost == c.IP {
		encode := fmt.Sprintf("%x", sha1.Sum([]byte(c.IP)))

		c.Host = fmt.Sprintf("%s-%s.%s.%s.%s.IP", c.Server.Config.HostPrefix,
			encode[0:5], encode[6:11], encode[12:17], encode[17:22])
	} else {
		pieces := strings.Split(c.RealHost, ".")
		str := fmt.Sprintf("%x", sha1.Sum([]byte(pieces[0])))
		str = str[0:10]
		if len(pieces) > 1 {
			for _, piece := range pieces[1:] {
				str = "." + piece
			}
		}
		c.Host = c.Server.Config.HostPrefix + "-" + str
	}
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

func (c *Client) Whois(target *Client) {
	var buf bytes.Buffer

	c.SendNumeric(RPL_WHOISUSER, target.Nick, target.Ident, target.Host, "*", target.Name+" ")
	c.SendNumeric(RPL_WHOISMODES, target.Nick, fmt.Sprintf("is using modes +%s", target.Modes.FlagString()))
	c.SendNumeric(RPL_WHOISHOST, target.Nick, fmt.Sprintf("is connecting from *@%s %s", target.RealHost, target.IP))

	c.Channels.lock.RLock()
	defer c.Channels.lock.RUnlock()

	for _, channel := range c.Channels.list {
		prefix := channel.ModePrefixFor(target)

		buf.WriteString(prefix + channel.Name)
		buf.WriteString(" ")
	}

	buf.WriteString(" ")
	if buf.Len() > 0 {
		c.SendNumeric(RPL_WHOISCHANNELS, target.Nick, buf.String())
	}

	c.SendNumeric(RPL_WHOISSERVER, target.Nick, target.Server.Name, target.Server.Network)

	// RPL_WHOISSECURE

	c.SendNumeric(RPL_ENDOFWHOIS, target.Nick, "End of /WHOIS list")
}

// sendMessage -
func (c *Client) sendMessage(msg *Message) error {
	_, err := c.sock.Write(msg.String())
	return err
}
