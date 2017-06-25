package ircd

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
	Command string   `json:"command"`
	Prefix  string   `json:"prefix"`
	Target  string   `json:"target"`
	Args    []string `json:"args"`

	rawLine string
}

// NewMessage returns a ptr to a message object given source and line
func NewMessage(source Messageable, line string) *Message {
	m := &Message{
		rawLine:    line,
		MessageSrc: source,
	}

	m.Parse()
	return m
}

// Parse raw line into the various attribtues of the Message struct
func (msg *Message) Parse() {

}
