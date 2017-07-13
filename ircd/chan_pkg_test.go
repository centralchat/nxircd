package ircd_test

import (
	"nxircd/ircd"
	"testing"
)

func TestChannel(t *testing.T) {
	server := NewTestServer()

	t.Run("it should add/remove a client", func(t *testing.T) {
		socket, cli := ircd.NewTestClient(server, "testy", "mcTester")

		ch := server.FindOrAddChan("#test")
		ch.Join(cli)

		if ch.Clients.Count() != 1 {
			t.Fatalf("ch.Clients: %d to be 1", ch.Clients.Count())
		}

		if cli.Channels.Count() != 1 {
			t.Fatalf("cli.Channels: %d to be 1", cli.Channels.Count())
		}
		// Check to make sure the proper thing was sent
		line := socket.GrabWriteLine()
		if line != ":testy JOIN #test" {
			t.Fatalf("Invalid line sent to client on join: (%s)", line)
		}

		line = socket.GrabWriteLine()
		if line != ":testing 324 testy #test +" {
			t.Fatalf("Did not send modes to client")
		}

		socket.GrabWriteLine()

		// burn 2 lines
		line = socket.GrabWriteLine()
		if line != ":testing 353 testy = #test :testy" {
			t.Fatalf("No names sent to client: %s", line)
		}

		// Ignore the end of names for now
		socket.GrabWriteLine()

		// Should Set mode +o
		line = socket.GrabWriteLine()
		if line != ":testing MODE #test +o testy" {
			t.Fatalf("No mode o set on empty channel: %s", line)
		}

		ch.Part(cli, "Leaving")

		if ch.Clients.Count() != 0 {
			t.Fatalf("ch.Clients: %d to be 0", ch.Clients.Count())
		}

		if cli.Channels.Count() != 0 {
			t.Fatalf("cli.Channels: %d to be 0", cli.Channels.Count())
		}

		line = socket.GrabWriteLine()
		if line != ":testy!mcTester@127.0.0.1 PART #test :Leaving" {
			t.Fatalf("Invalid part no ack to client: %s", line)
		}
	})

	t.Run("it should remove a user to the channel.", func(t *testing.T) {
		_, cli := ircd.NewTestClient(server, "testy", "mcTester")
		ch := server.FindOrAddChan("#test")
		ch.Join(cli)

	})
}
