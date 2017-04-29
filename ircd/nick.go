package ircd

import (
  "strings"
)

import "github.com/DanielOaks/girc-go/ircmsg"

//TODO

func nickCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  nick := string(msg.Params[0])
  nicknameRaw := strings.TrimSpace(nick)

  if !client.isRegistered {
    client.SetNick(nicknameRaw)
    return true
  }

  oldNick := client.nick

  client.ChangeNick(nicknameRaw)
  client.updateMasks()

  client.server.log.Info("Client changed nick: %s => %s", oldNick, client.nick)

  return true
}
