package ircd

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DanielOaks/girc-go/ircmsg"
)

type ChannelClientList map[string]*ChannelClient

type ChannelClient struct {
	modes  ModeList
	client *Client
}

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

	empty := channel.IsEmpty()

	channel.lock.Lock()
	channel.clients.Add(client)
	client.channels.Add(channel)
	channel.lock.Unlock()

	channel.Send(client.nickMask, "JOIN", channel.name)
	channel.SendTopicNumeric(client)
	channel.Names(client)

	if empty {
		// TODO: CLean up
		channel.SetMode(client, "o")
	}
}

func (channel *Channel) SetMode(client *Client, modeChar string) {
	mode := ChannelModes.Find(modeChar)
	cclient := channel.clients[client.nick]
	cclient.modes[modeChar] = *mode
	channel.Send(client.server.name, "MODE", channel.name, "+"+modeChar, client.nick, "")
}

// IsEmpty Returns true of the channel is empty and no clients are in it
// We should make this more robust for services going forward
func (channel *Channel) IsEmpty() bool {
	return (len(channel.clients) == 0)
}

// Send Names list to a user.
func (channel *Channel) Names(source *Client) {
	var buf bytes.Buffer

	for nick, _ := range channel.clients {
		buf.WriteString(nick)
		buf.WriteString(" ")
	}

	source.SendNumeric(RPL_NAMREPLY, "=", channel.name, buf.String())
	source.SendNumeric(RPL_ENDOFNAMES, channel.name, "End of /NAMES list")
}

// Part a client from a channel
func (channel *Channel) Part(client *Client, message string) {
	channel.Send(client.nickMask, "PART", channel.name, message)
	channel.Remove(client)
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

	for _, cclient := range channel.clients {
		cclient.client.socket.Write(line)
	}

	return
}

// SendTopic sends a TOPIC Numeric to a user.
func (channel *Channel) SendTopicNumeric(client *Client) {
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

	for nick, cclient := range channel.clients {
		if nick != prefix {
			cclient.client.socket.Write(line)
		}
	}

	return
}

func (ccl ChannelClientList) Add(client *Client) {
	ccl[client.nick] = &ChannelClient{
		client: client,
		modes:  make(ModeList),
	}
}

func (ccl ChannelClientList) Delete(client *Client) {
	delete(ccl, client.nick)
}
