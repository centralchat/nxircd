package ircd

import (
  "strings"
  "sync"
)

// ChannelListMap -
type ChannelListMap map[string]*Channel

// ChannelList stores channels and allows for easy lookup of them
type ChannelList struct {
  list ChannelListMap
  lock sync.RWMutex
}

// NewChannelList - Returns a new instance of a channel list
func NewChannelList() *ChannelList {
  return &ChannelList{
    list: make(ChannelListMap),
  }
}

// Add - Add a channel to our channel list
func (cl *ChannelList) Add(channel *Channel) {
  lowerName := strings.ToLower(channel.name)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  cl.list[lowerName] = channel
}

// Find a channel from our channel list
func (cl *ChannelList) Find(name string) (channel *Channel) {
  lowerName := strings.ToLower(name)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  channel = cl.list[lowerName]
  return
}

// DeleteByName -
func (cl *ChannelList) DeleteByName(name string) {
  lowerName := strings.ToLower(name)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  delete(cl.list, lowerName)
  return
}

// Delete a channel from the ChannelList
func (cl *ChannelList) Delete(channel *Channel) {
  lowerName := strings.ToLower(channel.name)
  cl.lock.Lock()
  defer cl.lock.Unlock()

  delete(cl.list, lowerName)
  return
}

// Move oldKey - delete it and place channel in new spot
func (cl *ChannelList) Move(oldKey string, channel *Channel) {
  cl.DeleteByName(oldKey)
  cl.Add(channel)
}
