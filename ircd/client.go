package ircd

// The reasoning behind this was so the web package could import and the services
// package can import without
import "nxircd/interfaces"
import "github.com/DanielOaks/girc-go/ircmsg"
import "fmt"
import "time"

const (
	capStateStart = iota
	capStateNeg   = iota
	capStateEnd   = iota
)

// Client Holds an IRC Client
type Client struct {
	Messageable

	sock interfaces.Socket

	Nick  string
	Ident string
	Name  string
	Host  string

	CTime time.Time
	ATime time.Time

	IP       string
	RealHost string
	HostMask string

	registered bool

	Server *Server

	local bool

	capState int
}

// NewClient Returns a new IRC Client
func NewClient(server *Server, sock interfaces.Socket) *Client {
	return &Client{
		sock:   sock,
		Server: server,
		local:  server.me,
		CTime:  time.Now(),
		ATime:  time.Now(),
	}
}

// Run our client loop
func (c *Client) Run() {
	var err error

	for err == nil {
		line, err := c.sock.Read()
		if err != nil {
			c.sock.Close()
			continue
		}
		msg, _ := ircmsg.ParseLineMaxLen(line, 512, 512)
		fmt.Println(msg)
	}
}

func (c *Client) Prefix() string {
	return ""
}

func (c *Client) Send(msg *Message) error {
	return nil
}
