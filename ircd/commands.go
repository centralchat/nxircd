package ircd

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

// Run - Perform the command
func (command *Command) Run(client *Client, msg ircmsg.IrcMessage) (cmdStatus bool) {
  if len(msg.Params) < command.minParams {
    return false
  }

  cmdStatus = command.handler(client, msg)
  return
}

//CommandList Holds the list of available commands
var CommandList = map[string]Command{
  "CAP": {
    handler:   capCmdHandler,
    minParams: 1,
  },
}
