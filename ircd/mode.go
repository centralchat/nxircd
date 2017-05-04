package ircd

import (
	"fmt"

	"github.com/DanielOaks/girc-go/ircmsg"
)

// Maybe dont need?
var ModeFlagOperator = 0x00001
var ModeFlagAdmin = 0x00002
var ModeFlagOwner = 0x00004
var ModeFlagHalfop = 0x00008
var ModeFlagVoice = 0x00010

type Mode struct {
	character  string
	symbol     string
	setparam   int
	unsetparam int
	flag       int
}

type ModeList map[string]Mode

type SupportedModes []Mode

type ModeAction struct {
	operator string
	mode     *Mode
	args     string
}

var ChannelModes = SupportedModes{
	{"q", "~", 1, 1, ModeFlagOwner},    // Channel Owner
	{"a", "&", 1, 1, ModeFlagAdmin},    // Channel Admin
	{"o", "@", 1, 1, ModeFlagOperator}, // Channel Operator
	{"h", "%", 1, 1, ModeFlagHalfop},   // Channel HalfOp
	{"v", "+", 1, 1, ModeFlagVoice},    // Channel Voice
}

func (sm *SupportedModes) parse(params ...string) (modes []*ModeAction) {
	operator := string(params[0][0])

	fmt.Println("Parsing modes: ", params)

	for pos, modChar := range params[0][1:] {
		fmt.Println("Processing ModeChar: ", string(modChar))
		mode := sm.Find(string(modChar))
		ma := &ModeAction{
			mode:     mode,
			operator: operator,
		}

		if operator == "+" && mode.setparam == 1 {
			ma.args = params[pos+1]
		}

		if operator == "-" && mode.unsetparam == 1 {
			ma.args = params[pos+1]
		}

		modes = append(modes, ma)
	}
	return
}

func (sm *SupportedModes) Find(mode string) *Mode {
	for _, m := range *sm {
		if m.character == mode {
			return &m
		}
	}
	return nil
}

func cmdModeHandler(client *Client, msg ircmsg.IrcMessage) bool {
	target := msg.Params[0]

	fmt.Println("Received Mode for: ", target, string(target[0]))
	if string(target[0]) == "#" {
		if len(msg.Params) > 1 {
			return processChannelMode(target, client, msg.Params[1:])
		}
	}

	return true
}

func processChannelMode(target string, client *Client, params []string) bool {
	channel := client.server.channels.Find(target)
	if channel == nil {
		client.SendNumeric(ERR_NOSUCHCHANNEL, target, "No such channel.")
		return false
	}

	actions := ChannelModes.parse(params...)
	for _, action := range actions {
		cli := channel.clients[action.args]
		mode := action.mode
		switch action.operator {
		case "+":
			cli.modes[mode.character] = *mode
		case "-":
			delete(cli.modes, mode.character)
			break
		}
	}

	parv := params
	parv = append([]string{channel.name}, parv...)

	channel.Send(client.nick, "MODE", parv...)
	return true
}
