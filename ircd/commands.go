package ircd

import (
	"fmt"
	"strings"
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
	client.Server.Log.DebugF("Running Command: %s ", msg.Command)

	if len(msg.Args) < cmd.minParams {
		client.Server.Log.DebugF("Not enough Arguments: %d < %d", len(msg.Args), cmd.minParams)
		return false
	}

	if err := cmd.handler(client.Server, client, msg); err != nil {
		// Do something with this err
		return false
	}
	return true
}

// TODO: Refactor this but for now put it here.
var clientCmdMap = map[string]ClientCmd{
	"NICK": {
		minParams: 1,
		handler:   nickUCmdHandler,
	},
	"USER": {
		minParams: 1,
		handler:   userUCmdHandler,
	},
	"QUIT": {
		minParams: 0,
		handler:   quitUCmdHandler,
	},
	"PART": {
		minParams: 1,
		handler:   partUCmdHandler,
	},
	"JOIN": {
		minParams: 1,
		handler:   joinUCmdHandler,
	},
	"WHO": {
		minParams: 1,
		handler:   whoUCmdHandler,
	},
	"PRIVMSG": {
		minParams: 2,
		handler:   msgUCmdHandler,
	},
	"NOTICE": {
		minParams: 2,
		handler:   noticeUCmdHandler,
	},
	"MODE": {
		minParams: 1,
		handler:   modeUCmdHandler,
	},
	"TOPIC": {
		minParams: 1,
		handler:   topicUCmdHandler,
	},
	"KICK": {
		minParams: 2,
		handler:   kickUCmdHandler,
	},
	"LIST": {
		handler: listUCmdHandler,
	},
}

func nickUCmdHandler(srv *Server, cli *Client, m *Message) error {
	nick := m.Args[0]

	if srv.NickInUse(nick) {
		// TODO Send no such nick
		cli.SendNumeric(ERR_NICKNAMEINUSE, nick, "*", "nick in use")
		return fmt.Errorf("nick in use")
	}

	if !ValidNick(nick) {
		// This is what unreal sends? cray
		cli.SendNumeric(ERR_NICKNAMEINUSE, nick, "*", "Nickname is unavailable: Illegal characters")
		return fmt.Errorf("invalid nick")
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

	srv.Log.InfoF("Client Connected: %s", cli.HostMask())

	return nil
}

func joinUCmdHandler(srv *Server, cli *Client, m *Message) error {
	names := strings.Split(m.Args[0], ",")
	if len(m.Args) > 1 {
		names = append(names, m.Args[1:]...)
	}

	for _, cName := range names {
		if err := joinChannel(cli, cName); err != nil {
			cli.SendNumeric(ERR_NOSUCHCHANNEL, cName, "*", "invalid channel")
		}
	}
	return nil
}

// WHO
//
// reply format:
// "<channel> <user> <host> <server> <nick> ( "H" / "G" > ["*"] [ ( "@" / "+" ) ] :<hopcount> <real name>" */
func whoUCmdHandler(srv *Server, cli *Client, m *Message) error {
	//	var operOnly bool

	// TODO: zero args should show all non-invisible users
	if len(m.Args) < 1 {
		cli.SendNumeric(RPL_ENDOFWHO, "end of /WHO")
	}

	target := m.Args[0]

	if ValidChannel(target) {
		ch := srv.FindChannel(target)

		if ch == nil {
			cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
			return fmt.Errorf("no such channel")
		}

		for _, cl := range ch.ClientsInChannel() {
			// TODO: Handle the H / G etc
			prefix := ch.ModePrefixFor(cl)
			cli.SendNumeric(RPL_WHOREPLY, target, cl.Ident, cl.Host, srv.Name, cl.Nick, "H"+prefix, fmt.Sprintf("0 %s", cl.Name))
		}
		cli.SendNumeric(RPL_ENDOFWHO, target, "end of /WHO")
		return nil

	}
	// WHO for nick
	if cl := srv.FindClient(target); cl != nil {
		cli.SendNumeric(RPL_WHOREPLY, "0", cl.Ident, cl.Host, srv.Name, cl.Nick, "H", fmt.Sprintf("0 %s", cl.Name))
	}

	cli.SendNumeric(RPL_ENDOFWHO, target, "end of /WHO")
	return nil
}

func quitUCmdHandler(srv *Server, cli *Client, m *Message) error {
	msg := m.Args[0]

	for _, ch := range cli.Channels.list {
		ch.Quit(cli, msg)
	}

	srv.Clients.Delete(cli)
	return nil
}

func partUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]

	if !ValidChannel(target) {
		cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "invalid channel")
		return fmt.Errorf("invalid channel")
	}

	ch := cli.Channels.Find(target)
	if ch == nil {
		cli.SendNumeric(ERR_NOTONCHANNEL, target, "*", "not on channel")
		return fmt.Errorf("not on channel")
	}

	msg := "left the channel"
	if m.Argc > 1 {
		msg = m.Args[m.Argc-1]
	}

	ch.Part(cli, msg)
	cli.Channels.Delete(ch)

	return nil
}

//TODO The Notice/PRIVMSG command look identical refactor to single thing
func msgUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]

	if ValidChannel(target) {
		ch := srv.FindChannel(target)
		if ch == nil {
			cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
			return fmt.Errorf("no such channel")
		}
		ch.PrivMsg(cli, m.Args[1])
		return nil
	}

	ct := srv.FindClient(target)
	if ct == nil {
		cli.SendNumeric(ERR_NOSUCHNICK, target, "*", "no such nick")
		return fmt.Errorf("no such nick")
	}

	ct.PrivMsg(cli, m.Args[1])
	return nil
}

func noticeUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]

	if ValidChannel(target) {
		ch := srv.FindChannel(target)
		if ch == nil {
			cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
			return fmt.Errorf("no such channel")
		}
		ch.Notice(cli, m.Args[1])
		return nil
	}

	ct := srv.FindClient(target)
	if ct == nil {
		cli.SendNumeric(ERR_NOSUCHNICK, target, "*", "no such nick")
		return fmt.Errorf("no such nick")
	}

	ct.Notice(cli, m.Args[0])
	return nil
}

func modeUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]
	fmt.Println("Target: ", target)

	if ValidChannel(target) {
		ch := srv.FindChannel(target)
		if ch == nil {
			cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
			return fmt.Errorf("no such channel")
		}

		// TODO: Add Access checks here
		changes := ParseCMode(m.Args[1:]...)
		ch.ApplyModeChanges(cli, changes)
		return nil
	}

	return nil
}

func topicUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]

	if !ValidChannel(target) {
		cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel: invalid channel name")
		return fmt.Errorf("invalid channel")
	}

	ch := srv.FindChannel(target)
	if ch == nil {
		cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
		return fmt.Errorf("no such channel")
	}

	if len(m.Args) == 1 {
		ch.sendTopicNumeric(cli)
		return nil
	}

	if !ch.IsOperator(cli) {
		cli.SendNumeric(ERR_NOPRIVS, target, "*", "no such channel")
		return fmt.Errorf("no privs")
	}
	ch.SetTopic(cli, m.Args[1])
	return nil
}

func kickUCmdHandler(srv *Server, cli *Client, m *Message) error {
	target := m.Args[0]
	if !ValidChannel(target) {
		cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel: invalid channel name")
		return fmt.Errorf("invalid channel")
	}

	ch := srv.FindChannel(target)
	if ch == nil {
		cli.SendNumeric(ERR_NOSUCHCHANNEL, target, "*", "no such channel")
		return fmt.Errorf("no such channel")
	}

	if !ch.IsHalfOp(cli) && !ch.IsOperator(cli) {
		cli.SendNumeric(ERR_NOPRIVILEGES, "kick", target, "no permissions.")
		return fmt.Errorf("no permissions")
	}

	client := ch.Clients.Find(m.Args[1])
	if client == nil {
		cli.SendNumeric(ERR_NOTONCHANNEL, target, m.Args[1], "no such nickname")
		return fmt.Errorf("no such nickname")
	}

	msg := "kicked from channel"
	if len(m.Args) > 2 {
		msg = m.Args[2]
	}

	ch.Kick(cli, client, msg)
	return nil
}

func listUCmdHandler(srv *Server, cli *Client, m *Message) error {
	srv.Channels.lock.RLock()
	defer srv.Channels.lock.RUnlock()

	for _, channel := range srv.Channels.list {
		cli.SendNumeric(RPL_LIST, channel.Name, fmt.Sprintf("%d", channel.Clients.Count()), fmt.Sprintf("[+%s] ", channel.Modes.FlagString()))
	}
	return nil
}
