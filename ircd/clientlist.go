package ircd

import (
	"fmt"
	"strings"
	"sync"
)

type clientListMap map[string]*Client

// ClientList stores clients and allows for easy lookup of them
type ClientList struct {
	list clientListMap
	lock sync.RWMutex
}

// NewClientList - Returns a new instance of a client list
func NewClientList() *ClientList {
	return &ClientList{
		list: make(clientListMap),
	}
}

// Add - Add a client to our client list
func (cl *ClientList) Add(client *Client) error {
	nick := strings.ToLower(client.Nick)
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if _, found := cl.list[nick]; found {
		return fmt.Errorf("nick in use")
	}

	cl.list[nick] = client
	return nil
}

// Find a client from our client list
func (cl *ClientList) Find(nick string) (client *Client) {
	lowerName := strings.ToLower(nick)
	cl.lock.Lock()
	defer cl.lock.Unlock()

	client = cl.list[lowerName]
	return
}

// DeleteByNick -
func (cl *ClientList) DeleteByNick(nick string) {
	lowerName := strings.ToLower(nick)
	cl.lock.Lock()
	defer cl.lock.Unlock()

	delete(cl.list, lowerName)
	return
}

// Delete a client from the ClientList
func (cl *ClientList) Delete(client *Client) {
	cl.DeleteByNick(client.Nick)
	return
}

// Move oldKey - delete it and place client in new spot
func (cl *ClientList) Move(oldKey string, client *Client) {
	cl.DeleteByNick(oldKey)
	cl.Add(client)
}

func (cl *ClientList) Count() int {
	return len(cl.list)
}
