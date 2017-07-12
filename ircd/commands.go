package ircd

import (
	"fmt"
)

type ClientCmd struct {
	handler      func(*Server, *Client, *Message) error
	requiresOper bool
	requiresReg  bool
	minParams    int
	capabs       []string
}

// Run - Perform the command
func (cmd *ClientCmd) Run(client *Client, msg *Message) (cmdStatus bool) {
	if len(msg.Args) < cmd.minParams {
		return false
	}

	if err := cmd.handler(client.Server, client, msg); err != nil {
		// Do something with this err
		return false
	}
	return true
}

// TODO: Refactor this but for now put it here.
var clientMsgTab = map[string]ClientCmd{
	"NICK": {
		minParams: 1,
		handler:   nickUCmdHandler,
	},
	"USER": {
		minParams: 1,
		handler:   userUCmdHandler,
	},
}

func nickUCmdHandler(srv *Server, cli *Client, m *Message) error {
	nick := m.Args[0]

	if srv.NickInUse(nick) {
		// TODO Send no such nick
		return fmt.Errorf("nick in use")
	}

	if !cli.registered {
		cli.SetNick(nick)
		return nil
	}

	cli.ChangeNick(nick)
	return nil
}

func userUCmdHandler(srv *Server, cli *Client, m *Message) error {
	if len(m.Args) != 4 {
		return fmt.Errorf("invalid arguments")
	}

	cli.Ident = m.Args[0]
	cli.Name = m.Args[3]

	srv.Greet(cli)
	srv.Log.InfoF("Client Connected: %s", cli.HostMask(true))

	return nil
}
