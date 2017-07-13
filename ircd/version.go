package ircd

import "fmt"

const VER_STRING = "0.0.1"
const VER_CODENAME = "Sunfire"

var VERSION = fmt.Sprintf("NxIRCD-%s(%s)", VER_STRING, VER_CODENAME)
