package ircd

import "sync"

type channelListMap map[string]*Channel

// ChanList stores the a list of channels
type ChanList struct {
	list channelListMap
	lock sync.RWMutex
}

func NewChanList() *ChanList {
	return &ChanList{
		list: make(channelListMap),
	}
}

func (cl *ChanList) Find(name string) *Channel {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	ch, _ := cl.list[name]
	return ch
}

func (cl *ChanList) Add(ch *Channel) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	cl.list[ch.Name] = ch
}

func (cl *ChanList) Delete(ch *Channel) {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	delete(cl.list, ch.Name)
}

func (cl *ChanList) Count() int {
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	return len(cl.list)
}
