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
    client.server.log.Info("Client Connected: %s [%s]", client.nick, client.ip)
    return true
  }

  oldNick := client.nick

  client.ChangeNick(nicknameRaw)
  client.server.log.Info("Client changed nick: %s => %s [%s]", oldNick, client.nick, client.ip)

  return true
}
