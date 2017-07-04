package ircd_test

import (
	"nxircd/ircd"
	"testing"
)

func TestClientList(t *testing.T) {
	t.Run("add a client", func(t *testing.T) {
		clientList := ircd.NewClientList()

		clientList.Add(&ircd.Client{Nick: "client1"})
		clientList.Add(&ircd.Client{Nick: "client2"})

		if client := clientList.Find("client1"); client == nil {
			t.Fatal("Client1 not found in the list after adding it")
		}
	})

	t.Run("should return error on dup add", func(t *testing.T) {
		clientList := ircd.NewClientList()

		if err := clientList.Add(&ircd.Client{Nick: "client1"}); err != nil {
			t.Fatal("should not error on first add")
		}

		if err := clientList.Add(&ircd.Client{Nick: "client1"}); err == nil {
			t.Fatal("should error on second add of dup")
		}
	})

	t.Run("remove a client by nick", func(t *testing.T) {
		clientList := ircd.NewClientList()

		clientList.Add(&ircd.Client{Nick: "client1"})

		clientList.DeleteByNick("client1")
		if clientList.Count() > 0 {
			t.Fatal("Client1 was not deleted from list by nick")
		}
	})

	t.Run("remove a client by ptr", func(t *testing.T) {
		clientList := ircd.NewClientList()

		clientList.Add(&ircd.Client{Nick: "client1"})

		client := clientList.Find("client1")
		if client == nil {
			t.Fatal("Client1 not found in the list after adding it")
		}

		clientList.Delete(client)
		if clientList.Count() > 0 {
			t.Fatal("Did not delete client from list")
		}
	})

	t.Run("Move a client in the list", func(t *testing.T) {
		clientList := ircd.NewClientList()

		client := &ircd.Client{Nick: "client1"}
		clientList.Add(client)

		client.Nick = "client2"
		clientList.Move("client1", client)

		if clientList.Find("client1") != nil {
			t.Fatal("client1 was not moved")
		}
		if clientList.Find("client2") == nil {
			t.Fatal("client1 was not moved")
		}

	})
}
