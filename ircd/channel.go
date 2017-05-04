package ircd

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DanielOaks/girc-go/ircmsg"
)

type ChannelClientList map[*Client]ModeList

type ChannelTopic struct {
	text   *string
	ctime  time.Time
	setter *Client
}

type Channel struct {
	name string

	topic *ChannelTopic

	server  *Server
	clients ChannelClientList
	lock    sync.RWMutex

	hasTopic bool
	modes    ModeList
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
	source.Send(source.server.name, RPL_ENDOFNAMES, channel.name, "End of /NAMES list")
}

// Part a client from a channel
func (channel *Channel) Part(client *Client, message string) {
	channel.Remove(client)
	channel.Send(client.nick, "PART", channel.name, message)
}

// Remove a client from a channel
func (channel *Channel) Remove(client *Client) {
	channel.lock.Lock()
	defer channel.lock.Unlock()

	channel.clients.Delete(client)
	client.channels.Delete(channel)
}

// Send a message to all users on a channel.
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

// SendTopic sends a TOPIC Numeric to a user.
func (channel *Channel) SendTopic(client *Client) {
	if channel.topic != nil {
		client.SendNumeric(RPL_TOPIC, channel.name, *channel.topic.text)
		client.SendNumeric(RPL_TOPICTIME, channel.name, channel.topic.setter.nick, strconv.FormatInt(channel.topic.ctime.Unix(), 10))
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
