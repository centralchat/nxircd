package ircd

import (
  "fmt"
  "strings"
)

import "github.com/DanielOaks/girc-go/ircmsg"

//TODO

func nickCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  nick := string(msg.Params[0])

  nicknameRaw := strings.TrimSpace(nick)

  fmt.Printf("Nick: %s\n", nicknameRaw)

  if !client.isRegistered {
    client.SetNick(nicknameRaw)
  }

  return true
}
