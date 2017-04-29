package ircd

import "fmt"

const VER_MAJOR = 0
const VER_MINOR = 1
const VER_PATCH = 0
const VER_BUILD = "-pre-alpha"

var VER_STRING = fmt.Sprintf("%d.%d.%d%s", VER_MAJOR, VER_MINOR, VER_PATCH, VER_BUILD)
