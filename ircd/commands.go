package ircd

type ClientCmd struct {
	handler      func(*Client, *Message) bool
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

	cmdStatus = cmd.handler(client, msg)
	return true
}

// TODO: Refactor this but for now put it here.
var clientMsgTab = map[string]ClientCmd{

	/**************************************/
	"NICK": {
		minParams: 1,
		// Nick Handler Start
		handler: func(cli *Client, m *Message) bool {
			nick := m.Args[0]

			if cli.Server.NickInUse(nick) {
				// TODO Send no such nick
				return false // No changes applied cmd failed
			}

			if !cli.registered {
				cli.SetNick(nick)
				return true
			}

			cli.ChangeNick(nick)
			return true
		},
		// Nick Handler END
	},

	/**************************************/
	"USER": {
		handler: func(cli *Client, m *Message) bool {
			return true
		},
	},
}
