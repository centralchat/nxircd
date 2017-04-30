package ircd

import (
  "bytes"
  "strings"
  "sync"
  "time"
)

import "github.com/DanielOaks/girc-go/ircmsg"

type ChannelClientList map[*Client]ModeList

type Channel struct {
  name string

  topic       string
  ttime       time.Time
  topicSetter *Client

  server  *Server
  clients ChannelClientList
  lock    sync.RWMutex

  hasTopic bool
  modes    ModeList
}

func NewChannel(name string, server *Server) *Channel {
  channel := &Channel{
    name:     name,
    server:   server,
    hasTopic: false,
    ttime:    time.Now(),
    clients:  make(ChannelClientList),
  }

  return channel
}

// Join a client to a channel
func (channel *Channel) Join(client *Client) {

  channel.lock.Lock()
  defer channel.lock.Unlock()

  channel.clients.Add(client)
  client.channels.Add(channel)

  channel.Send(client.nickMask, "JOIN", channel.name)

  channel.SendTopic(client)
  channel.Names(client)
}

func (channel *Channel) Names(source *Client) {
  var buf bytes.Buffer

  for client := range channel.clients {
    buf.WriteString(client.nick)
    buf.WriteString(" ")
  }

  source.Send(source.server.name, RPL_NAMREPLY, source.nick, "=", channel.name, buf.String())
  source.Send(source.server.name, RPL_ENDOFWHO, channel.name, "End of /Names list")
}

func (channel *Channel) Part(client *Client, message string) {
  channel.Remove(client)
  channel.Send(client.nick, "PART", channel.name, message)
}

func (channel *Channel) Remove(client *Client) {
  channel.lock.Lock()
  defer channel.lock.Unlock()

  channel.clients.Delete(client)
  client.channels.Delete(channel)
}

func (channel *Channel) Send(prefix string, command string, params ...string) (err error) {
  var line string

  message := ircmsg.MakeMessage(nil, prefix, command, params...)
  line, err = message.LineMaxLen(512, 512)
  if err != nil {
    channel.server.log.Warn("Send error %s\n", err)
    return
  }

  channel.server.log.Debug("--> %s\n", strings.TrimRight(line, "\r\n"))

  for client := range channel.clients {
    client.socket.Write(line)
  }

  return
}

func (channel *Channel) SendTopic(client *Client) {
  if channel.hasTopic {
    client.Send(client.server.name, RPL_TOPIC, channel.name, channel.topicSetter.nick, channel.topic)
    client.Send(client.server.name, RPL_TOPICTIME, channel.name, string(channel.ttime.Unix()))
  }
}

// SendToAllButPrefix TODO: Find a better way to handle this
func (channel *Channel) SendToAllButPrefix(prefix string, command string, params ...string) (err error) {
  var line string

  message := ircmsg.MakeMessage(nil, prefix, command, params...)
  line, err = message.LineMaxLen(512, 512)
  if err != nil {
    channel.server.log.Warn("Send error %s\n", err)
    return
  }

  channel.server.log.Debug("--> %s\n", strings.TrimRight(line, "\r\n"))

  for client := range channel.clients {
    if client.nick != prefix {
      client.socket.Write(line)
    }
  }

  return
}

func (ccl ChannelClientList) Add(client *Client) {
  ccl[client] = make(ModeList)
}

func (ccl ChannelClientList) Delete(client *Client) {
  delete(ccl, client)
}
