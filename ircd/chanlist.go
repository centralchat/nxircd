package ircd

import "sync"

type channelListMap map[string]*Client

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
