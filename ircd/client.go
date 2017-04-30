package ircd

// TODO: Move this to a package probable

import (
  "bytes"
  "crypto/sha256"
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
  vhost string

  name  string
  state int

  capVersion int

  realHost string
  ip       string

  ctime    time.Time
  atime    time.Time
  pingTime time.Time

  server *Server
  socket *Socket

  shouldStop   bool
  isRegistered bool
  useVhost     bool

  channels *ChannelList
  modes    ModeList

  // Masks
  nickMask string
  realMask string
}

/**************************************************************/

// NewClient - Create a new client
func NewClient(server *Server, conn net.Conn) *Client {
  var ip string

  now := time.Now()

  ip, _, _ = net.SplitHostPort(conn.RemoteAddr().String())

  client := &Client{
    ctime:    now,
    atime:    now,
    server:   server,
    ip:       ip,
    socket:   NewSocket(conn),
    state:    clientStateNew,
    channels: NewChannelList(),
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
      client.socket.closed = true
      client.Quit("Read error")
      break
    }
    client.server.log.Debug("<-- %s\n", line)

    // TODO: Handle this error
    msg, _ := ircmsg.ParseLineMaxLen(line, 512, 512)

    cmd, exists := CommandList[msg.Command]
    if !exists {
      if len(msg.Command) > 0 {
        client.Send(client.server.name, ERR_UNKNOWNCOMMAND, client.nick, msg.Command, "Unknown command")
      } else {
        client.Send(client.server.name, ERR_UNKNOWNCOMMAND, client.nick, "lastcmd", "No command given")
      }
      continue
    }

    _ = cmd.Run(client, msg)
  }

}

/**************************************************************/

func (client *Client) updateMasks() {
  client.nickMask = fmt.Sprintf("%s!%s@%s", client.nick, client.ident, client.host)
  client.realMask = fmt.Sprintf("%s!%s@%s", client.nick, client.ident, client.ip)
}

// super ugly but will improve it later
/**************************************************************/

func (client *Client) generateHostMask() string {
  var buffer bytes.Buffer

  h := sha256.New()
  h.Write([]byte(client.ip))
  buf := fmt.Sprintf("%x", h.Sum(nil))

  buffer.WriteString(buf[0:5])
  buffer.WriteString(".")
  buffer.WriteString(buf[6:11])
  buffer.WriteString(".")
  buffer.WriteString(buf[12:17])
  buffer.WriteString(".IP")

  return buffer.String()
}

/**************************************************************/

// Send a message to the client
// TODO: Implement tags
func (client *Client) Send(prefix string, command string, params ...string) (err error) {
  var line string

  message := ircmsg.MakeMessage(nil, prefix, command, params...)
  line, err = message.LineMaxLen(512, 512)
  if err != nil {
    client.server.log.Warn("Send error %s\n", err)
    return
  }

  client.server.log.Debug("--> %s\n", strings.TrimRight(line, "\r\n"))

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

func (source *Client) WhoReply(channel *Channel, client *Client) {
  channelName := channel.name
  flags := "H"

  source.Send(source.server.name, RPL_WHOREPLY, source.nick, channelName, client.ident, client.host, client.server.name, client.nick, flags, "0 "+client.name)
}

/**************************************************************/

// ChangeNick changes the nickname of a client
func (client *Client) ChangeNick(nick string) {
  //TODO: Add nick exists checks
  if !client.isRegistered {
    return
  }

  if cli := client.server.clients.Find(nick); cli != nil {
    client.Send(client.server.name, ERR_NICKNAMEINUSE, client.nick, fmt.Sprintf("%s is in use", nick))
    return
  }

  var oldNick = client.nick
  client.nick = nick

  client.server.clients.Move(oldNick, client)

  for _, cli := range client.CommonClients().list {
    cli.Send(oldNick, "NICK", client.nick)
  }
}

/**************************************************************/

// Quit remove a client from the network
func (client *Client) Quit(message string) {
  for _, channel := range client.channels.list {
    client.server.log.Debug("Removing: %s from %s", client.nick, channel.name)

    channel.Remove(client)                         // This locks
    channel.Send(client.nickMask, "QUIT", message) // This locks
  }
  client.server.clients.Delete(client)
  client.socket.Close()
}

/**************************************************************/

// Register - sets registered status on a client
func (client *Client) Register() {
  if client.isRegistered {
    return
  }

  client.host = client.generateHostMask()
  client.isRegistered = true
  client.state = clientStateReg
}

/**************************************************************/

// CommonClients are all the clients that are in channels with
// our user
func (client *Client) CommonClients() *ClientList {
  cl := NewClientList()

  cl.Add(client)

  for _, channel := range client.channels.list {
    channel.lock.RLock()
    for c := range channel.clients {
      cl.Add(c)
    }
    channel.lock.RUnlock()
  }
  return cl
}
