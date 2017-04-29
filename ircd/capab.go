package ircd

import (
  "fmt"
)
import "github.com/DanielOaks/girc-go/ircmsg"

// Rite now we support nothing
const ()

// A Capab String
type Capab string

// CapabSet a set of capabilities
type CapabSet map[Capab]bool

func capCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  client.state = clientStateCapStart
  server := client.server

  switch msg.Params[0] {
  case "LS":
    if !client.isRegistered {
      client.state = clientStateCapNeg
    }

    client.Send(server.name, "CAP", client.nick, "", "")

  case "END":
    if client.isRegistered {
      client.state = clientStateCapEnd
      client.server.register(client)
    }
  }
  fmt.Printf("%s\n", msg.Params[0])

  return true
}
