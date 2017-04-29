package ircd

import "bytes"

import (
  "github.com/DanielOaks/girc-go/ircmsg"
)

// Mode structure for modes
type Mode struct {
  mode      rune
  minParams int
  setable   bool
  handler   ModeApplier
}

const (
  Add    rune = '+'
  List   rune = '='
  Remove rune = '-'
)

type ModeChange struct {
  target    interface{}
  operation string
  mode      Mode
}

// ModeList Contains a list of modes
type ModeList map[*Mode]bool
type ModeApplier func(*Client, *ModeChange)

type SupportedModeList []Mode

var UserModes = SupportedModeList{
  {mode: 'i'},
  {mode: 'o'},
  {mode: 'O'},
  {mode: 'r'},
}

var ChannelModes = SupportedModeList{
  {mode: 'i'},
  {mode: 's'},
  {mode: 't'},
  {
    mode:      'b',
    minParams: 1,
  },
  {
    mode:      'l',
    minParams: 1,
  },
  {
    mode:      'q',
    minParams: 1,
  },
  {
    mode:      'o',
    minParams: 1,
  },
  {
    mode:      'h',
    minParams: 1,
  },
  {
    mode:      'v',
    minParams: 1,
  },
}

func (ml SupportedModeList) String() string {
  var buffer bytes.Buffer

  for _, m := range ml {
    buffer.WriteString(string(m.mode))
  }
  return buffer.String()
}

// Find a mode
func (ml SupportedModeList) Find(mode string) (md Mode) {
  for _, m := range ml {
    if string(m.mode) == mode {
      md = m
      return
    }
  }
  return
}

func modeCmdHandler(client *Client, msg ircmsg.IrcMessage) bool {
  return true
}
