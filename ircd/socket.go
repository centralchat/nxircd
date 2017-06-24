package ircd

import (
	"bufio"
	"io"
	"net"
)

// Socket Holds the basic methods for an ircd.Socket
// this allows us to override for non external clients
// like services etc
type Socket interface {
	Write(string) (int, error) // Write to the socket
	Read() (string, error)     // Read from the socket
	Close()
}

// IRCSocket extends Socket interface
// and holds information regarding a IRC Connection
type IRCSocket struct {
	Socket

	Closed bool

	conn net.Conn

	scanner *bufio.Scanner
	writer  *bufio.Writer
}

// NewIRCSocket takes a connection and returns a new ptr
// to an IRCSocket struct
func NewIRCSocket(conn net.Conn) *IRCSocket {
	return &IRCSocket{
		conn:    conn,
		Closed:  false,
		scanner: bufio.NewScanner(conn),
		writer:  bufio.NewWriter(conn),
	}
}

func (sock *IRCSocket) Write(line string) (int, error) {
	return sock.writer.WriteString(line)
}

func (sock *IRCSocket) Read() (string, error) {
	if sock.Closed {
		return "", io.EOF
	}
	return sock.scanner.Text(), sock.scanner.Err()
}
