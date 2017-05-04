package ircd

import (
	// For later
	_ "bytes"
	_ "strings"
	"time"

	"github.com/DanielOaks/girc-go/ircmsg"
)

// TODO: These commands should take a target with a pointer to the target
// if it should take an interface with the correct pointer being passed in
// expect channel commands should only work on channels
// with a handler to get them for everyone.

// For now getout cheep and use this one
// Since it has specs

// Command represents a command accepted from a client.
type Command struct {
	handler      func(client *Client, msg ircmsg.IrcMessage) bool
	requiresOper bool
	requiresReg  bool
	minParams    int
	capabs       []string
}

/************************************************************************************/

// Run - Perform the command
func (command *Command) Run(client *Client, msg ircmsg.IrcMessage) (cmdStatus bool) {
	if len(msg.Params) < command.minParams {
		return false
	}

	client.server.log.Debug("Running command: %s", msg.Command)

	cmdStatus = command.handler(client, msg)
	return
}

// TODO: Make this cleaner with this in a package
// Perhaps do command.RegisterCommand(&command.Command{
//   handler: capHandler
// })

/************************************************************************************/

//CommandList Holds the list of available commands
var CommandList = map[string]Command{
	// capab.go
	"CAP": {
		handler:   capCmdHandler,
		minParams: 1,
	},
	"NICK": {
		handler:   nickCmdHandler,
		minParams: 1,
	},
	"USER": {
		handler:   userCmdHandler,
		minParams: 1,
	},
	"PRIVMSG": {
		handler:   cmdPrivMsgHandler,
		minParams: 1,
	},
	"JOIN": {
		handler:   cmdJoinHandler,
		minParams: 1,
	},
	"PART": {
		handler:   cmdPartHandler,
		minParams: 1,
	},
	"TOPIC": {
		handler:   cmdTopicHandler,
		minParams: 1,
	},
	"WHO": {
		handler:   cmdWhoHandler,
		minParams: 1,
	},
	"WHOIS": {
		handler:   cmdWhoisHandler,
		minParams: 1,
	},
	"NAMES": {
		handler:   cmdNamesHandler,
		minParams: 1,
	},
	"MODE": {
		handler:   cmdModeHandler,
		minParams: 1,
	},
	"PING": {
		handler:   cmdPingHandler,
		minParams: 0,
	},
	"QUIT": {
		handler:   cmdQuitHandler,
		minParams: 1,
	},
}

// Find a better place for these

func userCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
	client.state = clientStateCapEnd
	// command := strings.ToUpper(string(msg.Params[0]))
	client.ident = msg.Params[0]
	client.name = msg.Params[3]

	client.server.register(client)

	client.updateMasks()

	client.server.log.Info("Client Connected: %s", client.realMask)

	return true
}

func cmdPrivMsgHandler(client *Client, msg ircmsg.IrcMessage) bool {
	var message = msg.Params[len(msg.Params)-1]
	for i, target := range msg.Params {
		if i > 3 {
			break
		}

		cli := client.server.clients.Find(target)
		if cli != nil {
			cli.Send(client.nick, "PRIVMSG", cli.nick, message)
			continue
		}

		channel := client.server.channels.Find(target)
		if channel != nil {
			channel.SendToAllButPrefix(client.nick, "PRIVMSG", channel.name, message)
		}
	}
	return true
}

func cmdJoinHandler(client *Client, msg ircmsg.IrcMessage) bool {
	channel := client.server.channels.Find(string(msg.Params[0]))
	if channel == nil {
		channel = NewChannel(msg.Params[0], client.server)
		client.server.channels.Add(channel)
	}

	channel.Join(client)
	return true
}

func cmdPartHandler(client *Client, msg ircmsg.IrcMessage) bool {
	channel := client.channels.Find(string(msg.Params[0]))

	if channel == nil {
		client.SendNumeric(ERR_NOTONCHANNEL, msg.Params[0], "You are not currently on the channel")
		return false
	}

	channel.Part(client, msg.Params[1])
	return true
}

func cmdWhoHandler(client *Client, msg ircmsg.IrcMessage) bool {
	// var buf bytes.Buffer

	channel := client.server.channels.Find(string(msg.Params[0]))
	if channel == nil {
		client.SendNumeric(ERR_NOSUCHCHANNEL, msg.Params[0], "No such channel.")
		return false
	}

	for _, ccli := range channel.clients {
		client.WhoReply(channel, ccli.client)
	}

	client.SendNumeric(RPL_ENDOFWHO, channel.name, "End of /WHO list")

	return true
}

func cmdWhoisHandler(client *Client, msg ircmsg.IrcMessage) bool {
	target := client.server.clients.Find(msg.Params[0])
	if target == nil {
		client.SendNumeric(ERR_NOSUCHNICK, msg.Params[0], "Client does not exist")
		return false
	}

	client.Whois(target)

	return true
}

func cmdPingHandler(client *Client, msg ircmsg.IrcMessage) bool {
	client.pingTime = client.server.time
	client.Send(client.server.name, "PONG", client.nick, msg.Params[0])
	return true
}

func cmdQuitHandler(client *Client, msg ircmsg.IrcMessage) bool {
	client.Quit(msg.Params[0])
	return true
}

func cmdNamesHandler(client *Client, msg ircmsg.IrcMessage) bool {
	channel := client.server.channels.Find(string(msg.Params[0]))
	if channel == nil {
		client.SendNumeric(ERR_NOSUCHCHANNEL, "No such channel.")
		return false
	}

	channel.Names(client)
	return true
}

func cmdTopicHandler(client *Client, msg ircmsg.IrcMessage) bool {
	channel := client.server.channels.Find(string(msg.Params[0]))
	if channel == nil {
		client.SendNumeric(ERR_NOSUCHCHANNEL, "No such channel.")
		return false
	}

	if len(msg.Params) == 1 {
		channel.SendTopicNumeric(client)
		return true
	}

	channel.topic = &ChannelTopic{
		text:   &msg.Params[1],
		ctime:  time.Now(),
		setter: client,
	}

	channel.lock.RLock()
	defer channel.lock.RUnlock()

	channel.Send(client.nick, "TOPIC", channel.name, *channel.topic.text)

	return true
}
