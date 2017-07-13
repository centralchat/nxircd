package ircd

import (
	"strings"
)

// Rite now we support nothing
const ()

// A Capab String
type Capab string

// CapabSet a set of capabilities
type CapabSet map[Capab]bool

func capCmdHandler(srv *Server, cli *Client, msg *Message) error {
	command := strings.ToUpper(msg.Args[0])

	switch command {
	case "LS":
		if !cli.registered {
			cli.capState = capStateStart
		}

		if len(msg.Args) > 1 && msg.Args[1] == "302" {
			cli.capVersion = 302
		}
		cli.SendFromServer("CAP", command, string(cli.capVersion))
	case "END":
		if cli.registered {
			cli.capState = capStateEnd
		}
	}
	return nil
}
