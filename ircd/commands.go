package ircd

import (
  _ "strings"
)

// For now getout cheep and use this one
// Since it has specs
import (
  "github.com/DanielOaks/girc-go/ircmsg"
)

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
}

/************************************************************************************/
// Find a better place for these

func userCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  client.state = clientStateCapEnd
  // command := strings.ToUpper(string(msg.Params[0]))
  client.ident = msg.Params[0]
  client.name = msg.Params[3]

  client.server.register(client)
  return true
}

/************************************************************************************/

func cmdPrivMsgHandler(client *Client, msg ircmsg.IrcMessage) bool {
  var message = msg.Params[len(msg.Params)-1]
  for i, target := range msg.Params {
    if i > 3 {
      break
    }

    target := client.server.clients.Find(target)
    if target != nil {
      target.Send(client.nick, "PRIVMSG", target.nick, message)
    }
  }
  return true
}
