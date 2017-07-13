package ircd

import (
	"regexp"
	"strings"
)

var (
	ChanExpr = regexp.MustCompile(`^[&!#+][\pL\pN]{1,63}$`)
	NickExpr = regexp.MustCompile("^[\\pL\\pN\\pP\\pS]{1,32}$")
)

// ValidChannel  IS the channel a valid name
func ValidChannel(name string) bool {
	return (strings.Index(name, "#") == 0 || strings.Index(name, "&") == 0)
}

func ValidNick(nick string) bool {
	if nick == "*" || strings.Contains(nick, ",") || strings.Contains("#@+", string(nick[0])) {
		return false
	}
	return NickExpr.MatchString(nick)
}
