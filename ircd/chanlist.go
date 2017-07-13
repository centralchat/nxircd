package ircd

import "sync"
import "strings"

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

func (cl *ChanList) Find(nameRaw string) *Channel {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	name := strings.ToLower(nameRaw)

	ch, _ := cl.list[name]
	return ch
}

func (cl *ChanList) Add(ch *Channel) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	name := strings.ToLower(ch.Name)

	cl.list[name] = ch
}

func (cl *ChanList) Delete(ch *Channel) {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	delete(cl.list, strings.ToLower(ch.Name))
}

func (cl *ChanList) Count() int {
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	return len(cl.list)
}
