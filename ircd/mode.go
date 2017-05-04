package ircd

import (
	"fmt"

	"github.com/DanielOaks/girc-go/ircmsg"
)

// Maybe dont need?

const (
	ModeFlagVoice    int = iota
	ModeFlagHalfop   int = iota
	ModeFlagOperator int = iota
	ModeFlagAdmin    int = iota
	ModeFlagOwner    int = iota
)

type Mode struct {
	character   string
	symbol      string
	setparam    int
	unsetparam  int
	flag        int
	minSetLevel int
}

type ModeList map[string]Mode

type SupportedModes []Mode

type ModeAction struct {
	operator string
	mode     *Mode
	args     string
}

var ChannelModes = SupportedModes{
	{"q", "~", 1, 1, ModeFlagOwner, ModeFlagOwner},       // Channel Owner
	{"a", "&", 1, 1, ModeFlagAdmin, ModeFlagOwner},       // Channel Admin
	{"o", "@", 1, 1, ModeFlagOperator, ModeFlagOperator}, // Channel Operator
	{"h", "%", 1, 1, ModeFlagHalfop, ModeFlagOperator},   // Channel HalfOp
	{"v", "+", 1, 1, ModeFlagVoice, ModeFlagHalfop},      // Channel Voice
}

func (ml *ModeList) highestSymbol() (prefix string) {
	mode := ml.highestMode()
	if mode != nil {
		return mode.symbol
	}
	return ""
}

func (ml *ModeList) highestMode() (mode *Mode) {
	mode = &Mode{}
	currentFlag := -1
	for _, m := range *ml {
		if m.flag > currentFlag {
			mode = &m
		}
	}
	return
}

func (sm *SupportedModes) parse(client *Client, params ...string) (modes []*ModeAction) {
	operator := string(params[0][0])

	fmt.Println("Parsing modes: ", params)

	for pos, modChar := range params[0][1:] {
		fmt.Println("Processing ModeChar: ", string(modChar))
		mode := sm.Find(string(modChar))
		if mode == nil {
			// Change the operator
			if modChar == '-' || modChar == '+' {
				operator = string(modChar)
				continue
			}

			client.SendNumeric(ERR_UNKNOWNMODE, string(modChar), "Unknown mode")
			return
		}

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

	cclient := channel.clients[client.nick]
	highestMode := cclient.modes.highestMode()

	actions := ChannelModes.parse(client, params...)
	for _, action := range actions {
		cli := channel.clients[action.args]
		mode := action.mode

		if mode.minSetLevel > highestMode.flag {
			client.SendNumeric(ERR_CHANOPRIVSNEEDED, channel.name,
				mode.character, "You do not have permissions to set this mode")
			return false
		}

		switch action.operator {
		case "+":
			cli.modes[mode.character] = *mode
			break
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
