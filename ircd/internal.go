package ircd

import (
	"fmt"
	"net"
)

// I do not like this here - perhaps we need an internal command handler
// that way we can keep this interaction out of this

func clientPreflight(c *Client) {
	c.SendFromServer("NOTICE", "AUTH", "*** Looking up your hostname.")

	names, err := net.LookupAddr(c.IP)
	if err != nil {
		c.SendFromServer("NOTICE", "AUTH", "*** Could not find your hostname; Using your ip instead.")
		c.RealHost = c.IP
	} else {
		c.SendFromServer("NOTICE", "AUTH", "*** Hostname found")
		c.RealHost = names[0]
	}

	c.Host = c.RealHost // For now to do masking later
}

func joinChannel(c *Client, cname string) error {
	if !ValidChannel(cname) {
		return fmt.Errorf("invalid channel name")
	}

	channel := c.Server.FindOrAddChan(cname)
	channel.Join(c)

	return nil
}
