package ircd

import (
  "strings"
)
import "github.com/DanielOaks/girc-go/ircmsg"

// Rite now we support nothing
const ()

// A Capab String
type Capab string

// CapabSet a set of capabilities
type CapabSet map[Capab]bool

func capCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  command := strings.ToUpper(msg.Params[0])

  client.state = clientStateCapStart

  server := client.server

  switch command {
  case "LS":
    if !client.isRegistered {
      client.state = clientStateCapNeg
    }

    if len(msg.Params) > 1 && msg.Params[1] == "302" {
      client.capVersion = 302
    }

    client.Send(server.name, "CAP", command, string(client.capVersion))
  case "END":
    if client.isRegistered {
      client.state = clientStateCapEnd
      client.server.register(client)
    }
  }

  return true
}
