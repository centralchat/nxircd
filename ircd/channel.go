package ircd

import (
  "strings"
  "sync"
)

import "github.com/DanielOaks/girc-go/ircmsg"

type ChannelClientList map[*Client]ModeList

type Channel struct {
  name string

  server  *Server
  clients ChannelClientList
  lock    sync.RWMutex

  modes ModeList
}

func NewChannel(name string, server *Server) *Channel {
  channel := &Channel{
    name:    name,
    server:  server,
    clients: make(ChannelClientList),
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
