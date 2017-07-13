package ircd_test

import (
	"nxircd/ircd"
	"testing"
)

func TestMessage(t *testing.T) {
	t.Run("should parse with ", func(t *testing.T) {
		m, err := ircd.NewMessage("PRIVMSG duder :Something something")
		if err != nil {
			t.Fatalf("Error when parsing message: %v", err)
		}

		if m.Prefix != "" {
			t.Fatal("Prefix should be blank")
		}

		if m.Command != "PRIVMSG" {
			t.Fatalf("Command doesnt == PRIVMSG its (%s)", m.Command)
		}

		if len(m.Args) != 2 {
			t.Fatalf("Not enough args shoudl be 2 instead: %d", len(m.Args))
		}
	})

	t.Run("should construct an IRC line correctly", func(t *testing.T) {
		m := map[string]*ircd.Message{
			"PRIVMSG Wanksta :Hey\r\n":       ircd.MakeMessage("", "PRIVMSG", "Wanksta", "Hey "),
			":Mitch PRIVMSG Wanky :Hey\r\n":  ircd.MakeMessage("Mitch", "PRIVMSG", "Wanky", "Hey "),
			"USER blah blah blah :Blah\r\n":  ircd.MakeMessage("", "USER", "blah", "blah", "blah", "Blah "),
			":Mitch JOIN #test\r\n":          ircd.MakeMessage("Mitch", "JOIN", "#test"),
			":Mitch PART #test :Leaving\r\n": ircd.MakeMessage("Mitch", "PART", "#test", "Leaving "),
		}

		for expected, msg := range m {
			if expected != msg.String() {
				t.Fatalf("Expected: \n\t %s\n to be \n\t%s", msg.String(), expected)
			}
		}
	})

	t.Run("should parse a msg with a prefix", func(t *testing.T) {
		m, err := ircd.NewMessage(":mandingo PRIVMSG duder :Something something")
		if err != nil {
			t.Fatalf("Error when parsing message: %v", err)
		}

		if m.Prefix != "mandingo" {
			t.Fatalf("Prefix was not properly set")
		}
	})

}
