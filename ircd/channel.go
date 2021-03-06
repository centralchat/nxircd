package ircd

import (
	"fmt"
	"strings"
	"time"
)

var ChanPrefix = "#"

var PermissionMap = map[string][]Mode{
	"kick": []Mode{'o', 'b'},
}

type Channel struct {
	Name string

	Clients *ClientList
	Modes   *ModeList

	// Topic Info
	Topic       string
	TopicSetter string
	TopicTime   time.Time

	CTime time.Time
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:    name,
		Clients: NewClientList(),
		Modes:   NewModeList(),
		CTime:   time.Now(),
	}
}

func (c *Channel) Join(cli *Client) {
	isNewChannel := c.Empty()

	cli.Reply("JOIN", c.Name)

	cli.SendNumeric(RPL_CHANNELMODEIS, c.Name, "+"+c.Modes.FlagString())
	cli.SendNumeric(RPL_CHANNELCREATED, c.Name, fmt.Sprintf("%d", c.CTime.Unix()))

	c.sendTopicNumeric(cli)

	c.Clients.Add(cli)
	cli.Channels.Add(c)

	c.Names(cli)

	if isNewChannel {
		c.AddModeServer(cli.Server, 'o', cli.Nick)
	} else {
		c.SendAllButPrefix(cli.HostMask(), "JOIN", c.Name)
	}
}

func (c *Channel) SetTopic(cli *Client, topic string) {
	c.TopicSetter = cli.Nick
	c.TopicTime = time.Now()
	c.Topic = topic

	c.Send(cli.HostMask(), "TOPIC", c.Name, c.Topic+" ")
}

func (c *Channel) sendTopicNumeric(cli *Client) {
	if c.Topic != "" {
		cli.SendNumeric(RPL_TOPIC, c.Name, c.Topic+" ")
		cli.SendNumeric(RPL_TOPICTIME, c.Name, c.TopicSetter, fmt.Sprintf("%d", c.TopicTime.Unix()))
	}
}

func (c *Channel) Names(cli *Client) {
	clients := c.ClientsInChannel()
	nicks := []string{}

	for _, client := range clients {
		prefix := c.ModePrefixFor(client)
		nicks = append(nicks, prefix+client.Nick)
	}

	cli.SendNumeric(RPL_NAMREPLY, "=", c.Name, strings.Join(nicks, " ")+" ")
	cli.SendNumeric(RPL_ENDOFNAMES, c.Name, "end of /NAMES	")
}

func (c *Channel) Empty() bool {
	return c.Clients.Count() == 0
}

func (c *Channel) IsOperator(cli *Client) bool {
	return c.Modes.HasExactArg('o', cli.Nick)
}

func (c *Channel) IsHalfOp(cli *Client) bool {
	return c.Modes.HasExactArg('h', cli.Nick)
}

func (c *Channel) IsVoice(cli *Client) bool {
	return c.Modes.HasExactArg('v', cli.Nick)
}

func (c *Channel) ClientCan(cli *Client, permName string) bool {
	modes, found := PermissionMap[permName]
	if !found {
		return true
	}

	for _, mode := range modes {
		if c.Modes.HasExactArg(mode, cli.Name) {
			return true
		}
	}
	return false
}

func (c *Channel) AllModePrefixesFor(cli *Client) string {
	var str string

	if c.IsOperator(cli) {
		str += "@"
	}
	if c.IsHalfOp(cli) {
		str += "%"
	}

	if c.IsVoice(cli) {
		str += "+"
	}

	return str
}

func (c *Channel) ModePrefixFor(cli *Client) string {
	if str := c.AllModePrefixesFor(cli); str != "" {
		return string(str[0])
	}
	return ""
}

// ApplyModeChanges to the current channel
// @params changes A list of mode changes and operations
func (c *Channel) ApplyModeChanges(setter *Client, changes []ModeChange) {
	if !c.CanSetModes(setter) {
		return // Reject all changes
	}

	for _, change := range changes {
		if c.IsHalfOp(setter) {
			if !change.Mode.IsAny('b', 'm', 'v') {
				// Reject the mode change
				continue
			}
		}

		switch change.Action {
		case ModeActionAdd:
			c.AddMode(setter, change.Mode, change.Arg)
		case ModeActionDel:
			c.DeleteMode(setter, change.Mode, change.Arg)
		case ModeActionList:
			// Ignore for now
			continue
		}
	}
}

func (c *Channel) CanSetModes(cli *Client) bool {
	return c.IsOperator(cli) || c.IsHalfOp(cli)
}

func (c *Channel) AddMode(setter *Client, m Mode, arg string) {
	// TODO: Propegate mode change effect
	c.Modes.Add(m, arg)
	c.Send(setter.HostMask(), "MODE", c.Name, "+"+m.String(), arg)
}

func (c *Channel) DeleteMode(setter *Client, m Mode, arg string) {
	// TODO: Propegate mode change effect
	c.Modes.DeleteArgument(m, arg)
	c.Send(setter.HostMask(), "MODE", c.Name, "-"+m.String(), arg)
}

func (c *Channel) AddModeServer(srv *Server, m Mode, arg string) {
	c.Modes.Add(m, arg)
	c.Send(srv.Name, "MODE", c.Name, "+"+m.String(), arg)
}

func (c *Channel) Part(cli *Client, msg string) {
	c.Send(cli.HostMask(), "PART", c.Name, msg+" ")
	c.Clients.Delete(cli)
	cli.Channels.Delete(c)
}

func (c *Channel) Kick(src *Client, cli *Client, msg string) {
	cli.Channels.Delete(c)

	// TODO: We really need a way to better mass set and unset modes
	c.Modes.DeleteArgument('o', cli.Nick)
	c.Modes.DeleteArgument('h', cli.Nick)
	c.Modes.DeleteArgument('v', cli.Nick)

	c.Send(src.HostMask(), "KICK", c.Name, cli.Nick, msg+" ")
	c.Clients.Delete(cli)
}

func (c *Channel) NickChange(cli *Client, oldNick string) {
	c.Clients.Move(oldNick, cli)
	c.Send(oldNick, "NICK", cli.Nick)
}

func (c *Channel) Quit(cli *Client, msg string) {
	c.Clients.Delete(cli)
	c.Send(cli.Nick, "QUIT", msg+" ")
}

func (c *Channel) PrivMsg(cli *Client, msg string) {
	c.SendAllButPrefix(cli.HostMask(), "PRIVMSG", c.Name, msg+" ")
}

func (c *Channel) Notice(cli *Client, msg string) {
	c.SendAllButPrefix(cli.HostMask(), "NOTICE", c.Name, msg+" ")
}

func (c *Channel) Send(prefix, cmd string, args ...string) {
	for _, client := range c.Clients.list {
		client.Send(prefix, cmd, args...)
	}
}

func (c *Channel) SendAllButPrefix(prefix, cmd string, args ...string) {
	for _, client := range c.Clients.list {
		if client.Nick != prefix && client.HostMask() != prefix {
			client.Send(prefix, cmd, args...)
		}
	}
}

func (c *Channel) ClientsInChannel() []*Client {
	arr := []*Client{}

	c.Clients.lock.RLock()
	defer c.Clients.lock.RUnlock()

	for _, client := range c.Clients.list {
		arr = append(arr, client)
	}
	return arr
}
