package ircd

import (
	"sync"
)

// Away          UserMode = 'a'
// Invisible     UserMode = 'i'
// LocalOperator UserMode = 'O'
// Operator      UserMode = 'o'
// Restricted    UserMode = 'r'
// ServerNotice  UserMode = 's' // deprecated
// WallOps       UserMode = 'w'

type Mode rune

func (m Mode) IsAny(modes ...Mode) bool {
	for _, mode := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

func (m Mode) String() string {
	return string(m)
}

type SupportedMode map[string]Mode

var SupportedUModes = SupportedMode{
	"away":       'a',
	"invisable":  'i',
	"localop":    'O',
	"globalop":   'o',
	"restricted": 'r',
	"service":    'S',
	"wallops":    'w',
	"bot":        'b',
	"snotice":    's',
}

var SupportedCModes = SupportedMode{
	"anonyous":   'a',
	"ban":        'b',
	"operonly":   'O',
	"optopic":    't',
	"key":        'k',
	"secret":     's',
	"registered": 'r',
	"limit":      'l',
	"moderated":  'm',
	"nooutside":  'n',

	// Perms
	"operator": 'o',
	"halfop":   'h',
	"voice":    'v',
}

func (sm SupportedMode) HasMode(m Mode) bool {
	for _, mo := range sm {
		if mo == m {
			return true
		}
	}
	return false
}

func (sm SupportedMode) String() string {
	str := ""
	for _, mode := range sm {
		str += mode.String()
	}
	return str
}

type ModeChange struct {
	Action Mode
	Mode   Mode
	Arg    string
}

var ModePrefixes = map[rune]rune{
	'o': '@',
	'h': '%',
	'v': '+',
}

// Set these as a mode cause we already have
// the string() method defined why waist types
const (
	ModeActionAdd  Mode = '+'
	ModeActionList Mode = '='
	ModeActionDel  Mode = '-'
)

// ModeList holds a list of either CModes or UModes
// for the server as well as any arguments associated with it
type ModeList struct {
	lock sync.RWMutex
	list map[Mode][]string
}

// NewModeList creates a modelist
func NewModeList() *ModeList {
	return &ModeList{
		list: make(map[Mode][]string),
	}
}
func (ml *ModeList) FlagString() string {
	modeStr := ""

	for mode, entry := range ml.list {
		if len(entry) == 0 {
			modeStr += mode.String()
		}
	}

	return modeStr
}

// Add a Mode and Arugment to the list
func (ml *ModeList) Add(m Mode, args ...string) {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	entry, found := ml.list[m]
	if !found {
		entry = []string{}
	}

	if args[0] != "" {
		entry = append(entry, args...)
	}

	ml.list[m] = entry
}

func (ml *ModeList) Delete(m Mode) {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	delete(ml.list, m)
}

func (ml *ModeList) DeleteArgument(m Mode, arg string) {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	entry, found := ml.list[m]
	if !found {
		return
	}

	ind := -1
	for pos, str := range entry {
		if str == arg {
			ind = pos
			break
		}
	}

	if ind == -1 {
		return
	}

	if ind == 0 {
		entry = entry[1:]
	} else if ind == len(entry) {
		entry = entry[0 : ind-1]
	} else {
		entry = append(entry, entry[0:ind-1]...)
		entry = append(entry, entry[ind+1:]...)
	}

	if len(entry) == 0 {
		delete(ml.list, m)
		return
	}
	ml.list[m] = entry
}

func (ml *ModeList) Find(m Mode) ([]string, bool) {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	entry, found := ml.list[m]
	return entry, found
}

// TODO: Refactor and optimize
func (ml *ModeList) HasExactArg(m Mode, arg string) bool {
	entry, found := ml.Find(m)
	if !found {
		return false
	}

	for _, marg := range entry {
		if marg == arg {
			return true
		}
	}
	return false
}

func ParseCMode(args ...string) []ModeChange {
	changes := []ModeChange{}

	for len(args) > 0 {
		modes := args[0]
		action := Mode(modes[0])

		switch action {
		case ModeActionAdd, ModeActionDel:
			modes = modes[1:]
		default:
			action = ModeActionList
		}

		pos := 1
		for _, mode := range modes {
			change := ModeChange{
				Action: action,
				Mode:   Mode(mode),
			}

			switch change.Mode {
			case 'k', 'b', 'i', 'l', 'o', 'h', 'v':
				if len(args) > pos {
					change.Arg = args[pos]
					pos++
				}
			}
			changes = append(changes, change)
		}
		args = args[pos:]
	}
	return changes
}
