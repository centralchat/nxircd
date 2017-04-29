package ircd

// TODO: Move this to a package probable

import (
  "fmt"
  "net"
  "strings"
  "time"
)

import "github.com/DanielOaks/girc-go/ircmsg"

/**************************************************************/

const (
  clientStateNew      = iota
  clientStateCapStart = iota
  clientStateCapNeg   = iota
  clientStateCapEnd   = iota
  clientStateReg      = iota
  clientStateDc       = iota
)

// Client - an IRCd Client
type Client struct {
  nick  string
  ident string
  host  string
  name  string
  state int

  capVersion int

  realHost   string
  remoteAddr net.Addr

  ctime time.Time
  atime time.Time

  server *Server
  socket *Socket

  shouldStop   bool
  isRegistered bool
}

/**************************************************************/

// NewClient - Create a new client
func NewClient(server *Server, conn net.Conn) *Client {
  now := time.Now()

  client := &Client{
    ctime:      now,
    atime:      now,
    server:     server,
    socket:     NewSocket(conn),
    state:      clientStateNew,
    remoteAddr: conn.RemoteAddr(),
  }

  client.run()

  return client
}

/**************************************************************/

func (client *Client) run() {
  var err error
  var line string

  for err == nil {
    if line, err = client.socket.Read(); err != nil {
      fmt.Printf("Socket error: %x", err)
      break
    }
    fmt.Printf("[%s] <-- %s\n", client.remoteAddr, line)

    // TODO: Handle this error
    msg, _ := ircmsg.ParseLineMaxLen(line, 512, 512)
    cmd, _ := CommandList[msg.Command]

    _ = cmd.Run(client, msg)
  }

}

/**************************************************************/

// Send a message to the client
// TODO: Implement tags
func (client *Client) Send(prefix string, command string, params ...string) (err error) {
  var line string

  message := ircmsg.MakeMessage(nil, prefix, command, params...)
  line, err = message.LineMaxLen(512, 512)
  if err != nil {
    fmt.Printf("Send %s\n", err)
    return
  }

  fmt.Printf("-> %s\n", strings.TrimRight(line, "\r\n"))

  client.socket.Write(line)
  return
}

/**************************************************************/

// Reply send a string back to the client
func (client *Client) Reply(reply string) error {
  return client.socket.Write(reply)
}

/**************************************************************/

// SetNick sets for the first time a clients nick
func (client *Client) SetNick(nick string) {
  //TODO: Add nick exists checks
  if !client.isRegistered {
    client.nick = nick
    client.server.clients.Add(client)
  }
}

/**************************************************************/

// Register - sets registered status on a client
func (client *Client) Register() {
  if client.isRegistered {
    return
  }
  client.isRegistered = true
  client.state = clientStateReg
}
