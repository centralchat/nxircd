package ircd

import (
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
func (cl *ClientList) Add(client *Client) {
  lowerName := strings.ToLower(client.nick)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  cl.list[lowerName] = client
}

// Get a client from our client list
func (cl *ClientList) Get(nick string) (client *Client) {
  lowerName := strings.ToLower(nick)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  client = cl.list[lowerName]
  return
}
