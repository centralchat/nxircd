package ircd

import (
	"bytes"
	"fmt"
	"strings"
)

type Messageable interface {
	// Should implement send
	Send(*Message) error

	// Must have a method prefix that returns nick!ident@host or
	// blah.servername.com
	Prefix() string

	// Target should return the following
	// nick for Client
	// name for Channel
	// name for Server
	Target() string
}

// Message holds our IRC Message
type Message struct {
	// Named like this so we can export Target string to json
	MessageSrc Messageable // Holds the source Struct
	// Not sure if im keeping this here as of right now,
	// might need to jsut make this a helper method to look it up
	MessageTrg Messageable // Holds the target Struct

	// Exportable JSON attributes so we can integrate CCIRC
	Command string `json:"command"`
	Prefix  string `json:"prefix"`

	// TODO: Figure out this
	// Target  string   `json:"target"`
	Args []string `json:"args"`

	rawLine string

	Blank bool `json:"blank"`
}

// NewMessage returns a ptr to a message object given source and line
func NewMessage(line string) (*Message, error) {
	m := &Message{}
	if len(line) == 0 {
		m.Blank = true
		return m, nil
	}

	str := strings.TrimSpace(line)
	if str[0] == ':' {
		p := strings.SplitN(str, " ", 2)
		if len(p) <= 1 {
			return m, fmt.Errorf("line to short")
		}
		m.Prefix = p[0][1:]
		str = p[1]
	}

	p := strings.SplitN(str, ":", 2)
	args := strings.Split(strings.TrimSpace(p[0]), " ")

	m.Command = strings.ToUpper(args[0])
	if len(args) > 1 {
		m.Args = append([]string{}, args[1:]...)
	}

	if len(p) > 1 {
		m.Args = append(m.Args, p[1])
	}

	return m, nil
}

// MakeMessage - TODO Make more logic here
func MakeMessage(prefix, cmd string, args ...string) *Message {
	return &Message{
		Command: cmd,
		Prefix:  prefix,
		Args:    args,
	}
}

func (m *Message) String() string {
	var buf bytes.Buffer

	if m.Prefix != "" {
		buf.WriteString(fmt.Sprintf(":%s ", m.Prefix))
	}

	buf.WriteString(m.Command)

	argc := len(m.Args)
	if argc > 0 {
		for pos, val := range m.Args {
			buf.WriteString(" ")
			if pos == argc-1 {
				buf.WriteString(":")
			}
			buf.WriteString(val)
		}
	}

	return buf.String()
}
